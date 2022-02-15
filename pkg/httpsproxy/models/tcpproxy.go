/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
	"github.com/theotw/natssync/pkg/httpsproxy/nats"
)

const (
	maxBytesToReadTCPEnvVariableKey = "MAX_BYTES_TCP"
	defaultMaxBytesToReadTCP        = 1024
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

//takes in a JSON byte array of the TcpData and returns the byte data
func DecodeTCPData(data []byte) ([]byte, error) {
	var tcpData TCPData
	err := json.Unmarshal(data, &tcpData)
	if err != nil {
		return nil, err
	}
	bits, err := base64.StdEncoding.DecodeString(tcpData.DataB64)
	if err != nil {
		return nil, err
	}
	if len(bits) != tcpData.DataLength {
		return nil, errors.New("mismatch data len ")
	}
	return bits, err
}

func StartBiDiNatsTunnel( outBoundSubject, inBoundSubject, connectionID string,inBoundQueue nats.NatsSubscriptionInterface, socket io.ReadWriteCloser) error {

	defer func() {
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
	log.Debugf("BIDI connection started  %s <-> %s ", outBoundSubject, inBoundSubject)
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
			if err := nc.Flush(); err != nil {
				log.WithError(err).Errorf("failed to flush natssync")
			}

			if writeErr != nil {
				log.WithError(writeErr).Errorf("Error writing data to nats")
				break
			} else {
				log.WithField("subject", subject).Debugf("Sent socket data to nats")
			}
		}

		if readErr != nil {
			errorString := readErr.Error()
			if !(strings.Contains(errorString,"EOF") || strings.Contains(errorString,"use of closed network connection")){
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
	for {
		log.Debug("waiting for Data from nats")
		natsMsg, err := queue.NextMsg(10 * time.Minute)
		if err != nil {
			log.WithError(err).Errorf("Error reading from NATS")
		} else {
			log.Debug("Got package from nats")
			tcpData, readErr := DecodeTCPData(natsMsg.Data)
			if readErr == nil {
				log.Debugf("Got valid package from nats len %d", len(tcpData))
				if len(tcpData) > 0 {
					if _, err := dest.Write(tcpData); err != nil {
						log.WithError(err).Errorf("failed to write tcp data to socket")
					}
				} else {
					//if we got 0 length data, we are done, bail
					break
				}
			} else {
				log.WithError(readErr).Error("Error reading data nats->tcp")
				break
			}
		}
	}
	log.Debug("Terminating")
	//send terminate
}
