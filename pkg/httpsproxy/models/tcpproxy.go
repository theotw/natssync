/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
	"github.com/theotw/natssync/pkg/httpsproxy/nats"
)

const (
	maxBytesToReadTCPEnvVariableKey = "MAX_BYTES_TCP"
	defaultMaxBytesToReadTCP        = 64 * 1024
)

func EncodeTCPData(data []byte, connectionID string, sequenceID int) []byte {
	encoded := base64.StdEncoding.EncodeToString(data)
	tcpData := &TCPData{
		DataB64:      encoded,
		DataLength:   len(data),
		SequenceID:   sequenceID,
		ConnectionID: connectionID,
	}
	bits, _ := json.Marshal(tcpData)
	return bits
}

// takes in a JSON byte array of the TcpData and returns the byte data
func DecodeTCPData(data []byte) ([]byte, int, error) {
	var tcpData TCPData
	err := json.Unmarshal(data, &tcpData)
	if err != nil {
		return nil, -1, err
	}
	bits, err := base64.StdEncoding.DecodeString(tcpData.DataB64)
	if err != nil {
		return nil, -1, err
	}
	if len(bits) != tcpData.DataLength {
		return nil, -1, errors.New("mismatch data len ")
	}
	return bits, tcpData.SequenceID, err
}

func StartBiDiNatsTunnel(outBoundSubject, inBoundSubject, connectionID string, inBoundQueue nats.NatsSubscriptionInterface, socket io.ReadWriteCloser) error {
	defer func() {
		log.Debugf("BIDI connection ended  connID %s %s <-> %s ", connectionID, outBoundSubject, inBoundSubject)
		if err := socket.Close(); err != nil {
			log.WithError(err).
				WithFields(
					log.Fields{
						"outBoundSubject": outBoundSubject,
						"inBoundSubject":  inBoundSubject,
						"connectionID":    connectionID,
					},
				).Errorf("failed to close socket")
		}
	}()

	//kick off the outbound stream tcp in to Nats
	go TransferTcpDataToNats(outBoundSubject, connectionID, socket)
	log.Debugf("BIDI connection started  connID %s %s <-> %s ", connectionID, outBoundSubject, inBoundSubject)
	//now read inbound NATS and write it back to the socket
	TransferNatsToTcpData(inBoundQueue, socket)
	return nil
}

func TransferTcpDataToNats(subject string, connectionID string, src io.ReadCloser) {
	nc := GetNatsClient()

	maxBytesToReadString := httpproxy.GetEnvWithDefaults(
		maxBytesToReadTCPEnvVariableKey,
		strconv.Itoa(defaultMaxBytesToReadTCP),
	)

	maxBytesToRead, err := strconv.Atoi(maxBytesToReadString)
	if err != nil {
		log.WithError(err).
			WithField(maxBytesToReadTCPEnvVariableKey, maxBytesToReadString).
			Errorf(
				"failed to get max bytes to read from environment, setting to default value of %v",
				defaultMaxBytesToReadTCP,
			)

		maxBytesToRead = defaultMaxBytesToReadTCP
	}

	sequenceID := 0
	for {
		log.Debug("Reading Data from socket")
		buf := make([]byte, maxBytesToRead)
		bufferLen, readErr := src.Read(buf)
		log.Debugf("Read %d bytes ", bufferLen)
		if bufferLen > 0 {
			writeBuf := buf[:bufferLen]
			sequenceID = sequenceID + 1
			dataToSend := EncodeTCPData(writeBuf, connectionID, sequenceID)

			writeErr := nc.Publish(subject, dataToSend)
			if writeErr != nil {
				log.WithError(writeErr).Errorf("Error writing data to nats")
				break
			} else {
				log.WithField("subject", subject).Debugf("Sent socket data to nats")
			}
			if err := nc.Flush(); err != nil {
				log.WithError(err).Errorf("failed to flush natssync")
			}
		}

		if readErr != nil {
			errorString := readErr.Error()
			if !(strings.Contains(errorString, "EOF") || strings.Contains(errorString, "use of closed network connection")) {
				log.WithError(readErr).Errorf("Error reading data tcp -> nats")
			}
			break
		}
	}

	writebuf := make([]byte, 0)
	sequenceID = sequenceID + 1
	dataToSend := EncodeTCPData(writebuf, connectionID, sequenceID)

	//send one last 0 len data package to send the stream
	log.WithField("subject", subject).Debug("Sent final packet data to nats")
	if err := nc.Publish(subject, dataToSend); err != nil {
		log.WithError(err).Errorf("failed to publish data")
	}
	if err := nc.Flush(); err != nil {
		log.WithError(err).Errorf("failed to flush natssync")
	}

	log.Debug("Terminating")
	//send terminate
}

func TransferNatsToTcpData(queue nats.NatsSubscriptionInterface, dest io.WriteCloser) {
	startTime := time.Now()

	sequenceIDOutgoing := 0
	emptyMessageArrivedTooEarly := false
	usedCache := false
	msgCache := make(map[int][]byte, 0)

	messageComplete := false
	for !messageComplete {
		log.Debugf("waiting for Data from nats")
		natsMsg, err := queue.NextMsg(1 * time.Minute)
		if err != nil {
			if !strings.Contains(err.Error(), "nats: timeout") {
				log.WithError(err).Errorf("Error reading from NATS")
				break
			}
			if time.Since(startTime).Minutes() > 3 {
				log.WithError(err).Errorf("Error reading from NATS -- exceeded 3 minutes, giving up now (empty msg arrived too early: %t)", emptyMessageArrivedTooEarly)
				break
			}
			// this is NextMsg timeout, we can try again
			continue
		}
		strickCheck := os.Getenv("STRICT_CONNECTION_CHECK")
		log.Debugf("Stick Check \"%s\"", strickCheck)

		if strickCheck == "true" {
			connectionID := natsMsg.Header.Get("x-connection-id")
			if len(connectionID) == 0 {
				log.WithError(err).Errorf("Not Connection ID header found on message")
				continue
			} else {
				log.Infof("Processing message for connection ID %s", connectionID)
			}
		}
		log.Debug("Got package from nats")
		tcpData, sequenceID, readErr := DecodeTCPData(natsMsg.Data)
		if readErr != nil {
			log.WithError(readErr).Error("Error reading data nats->tcp")
			break
		}

		if sequenceIDOutgoing+1 == sequenceID {
			// sequenceID is correct
			tcpDataLen := len(tcpData)
			log.Debugf("Got valid package from nats len %d", tcpDataLen)
			if tcpDataLen == 0 {
				//if we got 0 length data, we are done, bail
				if usedCache {
					log.Info("Message complete, but had to use cache")
				}
				break
			}

			sequenceIDOutgoing++
			if _, err := dest.Write(tcpData); err != nil {
				log.WithField("seconds", int(time.Since(startTime).Seconds())).WithError(err).Errorf("failed to write tcp data to socket %d", tcpDataLen)
				break // no need to continue
			}

			// continue reading from nats to get next message
			continue
		}

		// message came too early
		log.Errorf("Expected seq id %d, but got %d", sequenceIDOutgoing+1, sequenceID)

		if len(tcpData) == 0 {
			emptyMessageArrivedTooEarly = true
		}

		// keep message in cache for now
		msgCache[sequenceID] = tcpData
		usedCache = true

		// check if we have cached messages we can send now
		sendCachedMessagesCount := 0
		for true {
			nextOutgoing := sequenceIDOutgoing + 1
			tcpData, ok := msgCache[nextOutgoing]
			if !ok {
				// we do not have the next message, we will have to go back and wait for nats to send it
				break
			}
			sendCachedMessagesCount++

			tcpDataLen := len(tcpData)
			if tcpDataLen == 0 {
				//if we got 0 length data, we are done, bail
				messageComplete = true
				log.Info("Message complete, but had to use cache")
				break
			}

			if _, err := dest.Write(tcpData); err != nil {
				log.WithField("seconds", int(time.Since(startTime).Seconds())).WithError(err).Errorf("failed to write tcp data to socket %d", tcpDataLen)
			}
			sequenceIDOutgoing = nextOutgoing
			delete(msgCache, nextOutgoing)
		}

		if sendCachedMessagesCount > 0 {
			log.Infof("just sent %d cached messages...", sendCachedMessagesCount)
		}
	}
	log.Debug("Terminating")
	//send terminate
}
