/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg/httpsproxy"
	"io"
	"net"
	"time"
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

// HandleConnectionRequest is used in the proxlet to handle a connection request to a service in the internal network
// We attempt to estabilsh a socket to the target host:port.  If we can connect, we then setup two channels.  One listener for
// packets to write to a socket and then a publisher to send packets read from the socket
func HandleConnectionRequest(msg *nats.Msg, targetLocationID string) {
	var connectionMsg TCPConnectRequest
	var connectionResp TCPConnectResponse
	err := json.Unmarshal(msg.Data, &connectionMsg)
	if err != nil {
		log.Errorf("Error decing connection request %s", err.Error())
		connectionResp.State = "failed"
		connectionResp.StateDetails = err.Error()
	} else {
		targetSocket, err := net.DialTimeout("tcp", connectionMsg.Destination, 10*time.Second)

		if err != nil {
			log.Errorf("Error dialing connection to %s ", connectionMsg.Destination)
			connectionResp.State = "failed"
			connectionResp.StateDetails = err.Error() + " @ " + connectionMsg.Destination
		} else {
			connectionResp.State = "ok"
			outBoundSubject := httpproxy.MakeHttpsMessageSubject( connectionMsg.ProxyLocationID, connectionMsg.ConnectionID)
			inBoundSubject := httpproxy.MakeHttpsMessageSubject( targetLocationID, connectionMsg.ConnectionID)
			go StartBiDiNatsTunnel(outBoundSubject, inBoundSubject, connectionMsg.ConnectionID, targetSocket)
		}
	}
	nc := GetNatsClient()
	respbits, jsonerr := json.Marshal(&connectionResp)
	if jsonerr == nil {
		nc.Publish(msg.Reply, respbits)
		nc.Flush()
	} else {
		log.WithError(err).Error("Error marshaling connection resp message")
	}
}

func SendConnectionRequest(connectionID, clientID, host string) error {
	nc := GetNatsClient()
	reply := httpproxy.MakeReplyMessageSubject()
	sub := httpproxy.MakeMessageSubject(clientID, httpproxy.HTTPS_PROXY_CONNECTION_REQUEST)
	sync, err := nc.SubscribeSync(reply)
	if err != nil {
		log.Errorf("Error connecting to NATS subject  %s", err.Error())
		return err
	}
	var connectionMsg TCPConnectRequest
	connectionMsg.ConnectionID = connectionID
	connectionMsg.Destination = host
	connectionMsg.ProxyLocationID = httpproxy.GetMyLocationID()
	jsonBits, jsonerr := json.Marshal(&connectionMsg)
	if jsonerr != nil {
		return jsonerr
	}

	//_ = sync.AutoUnsubscribe(1)
	err = nc.PublishRequest(sub, reply, jsonBits)
	if err != nil {
		log.Errorf("Error Sendsing to NATS message  %s", err.Error())
		return err
	}

	err = nc.Flush()
	if err != nil {
		log.Errorf("Error flushing NATS  %s", err.Error())
		return err
	}

	respmsg, nextErr := sync.NextMsg(1 * time.Minute)
	if nextErr != nil {
		log.Errorf("Error reading nats msg %s", nextErr.Error())
		return nextErr
	}
	var resp TCPConnectResponse
	jsonerr = json.Unmarshal(respmsg.Data, &resp)
	if jsonerr != nil {
		return jsonerr
	}
	var respError error
	if resp.State != "ok" {
		respError = errors.New(resp.State + resp.StateDetails)
	}
	log.Debugf("End Connection request status: %s  details %s", resp.State, resp.StateDetails)

	return respError
}

func StartBiDiNatsTunnel(outBoundSubject, inBoundSubject, connectionID string, socket io.ReadWriteCloser) error {
	nc := GetNatsClient()
	defer socket.Close()

	//First, setup and subscribe to the inbound Subject
	inBoundQueue, err := nc.SubscribeSync(inBoundSubject)
	if err != nil {
		return err
	}
	//kick off the outbound stream tcp in to Nats
	go TransferTcpDataToNats(outBoundSubject, connectionID, socket)
	log.Debugf("BIDI connection started  %s <-> %s ", outBoundSubject, inBoundSubject)
	//now read inbound NATS and write it back to the socket
	TransferNatsToTcpData(inBoundQueue, socket)
	return nil
}

func TransferTcpDataToNats(subject string, connectioID string, src io.ReadCloser) {
	nc := GetNatsClient()

	sequnceID := 0
	for {
		log.Debug("Reading Data from socket")
		buf := make([]byte, 1024)
		len, readErr := src.Read(buf)
		log.Debugf("Read %d bytes ", len)
		if len > 0 {
			writebuf := buf[:len]
			sequnceID = sequnceID + 1
			dataToSend := EncodeTCPData(writebuf, connectioID, sequnceID)

			writeErr := nc.Publish(subject, dataToSend)
			nc.Flush()
			if writeErr != nil {
				log.Errorf("Error writing data to nats %s", writeErr.Error())
				break
			} else {
				log.Debugf("Sent socket data to nats %s", subject)
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				log.WithError(readErr).Errorf("Error reading data tcp -> nats")
			}
			break
		}
	}
	writebuf := make([]byte, 0)
	sequnceID = sequnceID + 1
	dataToSend := EncodeTCPData(writebuf, connectioID, sequnceID)
	//send one last 0 len data package to send the stream
	log.WithField("subject", subject).Debug("Sent final packet data to nats")
	nc.Publish(subject, dataToSend)
	nc.Flush()

	log.Debug("Terminating")
	//send terminate
}

func TransferNatsToTcpData(queue *nats.Subscription, dest io.WriteCloser) {
	for {
		log.Debug("waiting for Data from nats")
		natsMsg, err := queue.NextMsg(10 * time.Minute)
		if err != nil {
			log.Errorf("Error reading from NATS %s", err.Error())
		} else {
			log.Debug("Got package from nats")
			tcpData, readErr := DecodeTCPData(natsMsg.Data)
			if readErr == nil {
				log.Debugf("Got valid package from nats len %d", len(tcpData))
				if len(tcpData) > 0 {
					dest.Write(tcpData)
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
