/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg/httpsproxy"
	"github.com/theotw/natssync/pkg/httpsproxy/models"
)

type JsonError struct {
	Message string `json:"message"`
}

func HandleError(_ *gin.Context, err error) (int, interface{}) {
	ret := JsonError{
		Message: err.Error(),
	}

	return http.StatusInternalServerError, &ret
}

func (s *server) connectHandler(c *gin.Context) {
	s.connectHandlerNats(c)
	//connectHandlerLocal(c)
}

func (s *server) sendConnectionRequest(connectionID, clientID, host string) error {

	reply := httpproxy.MakeReplyMessageSubject(s.locationID, s.test)
	sub := httpproxy.MakeMessageSubject(clientID, httpproxy.HTTPS_PROXY_CONNECTION_REQUEST)
	sync, err := s.natsClient.SubscribeSync(reply)
	if err != nil {
		log.WithError(err).Error("Error connecting to NATS subject")
		return err
	}

	var connectionMsg models.TCPConnectRequest
	connectionMsg.ConnectionID = connectionID
	connectionMsg.Destination = host
	connectionMsg.ProxyLocationID = s.locationID
	jsonBits, jsonErr := json.Marshal(&connectionMsg)
	if jsonErr != nil {
		return jsonErr
	}

	if err := s.natsClient.PublishRequest(sub, reply, jsonBits); err != nil {
		log.WithError(err).Error("Error Sending NATS message")
		return err
	}

	if err := s.natsClient.Flush(); err != nil {
		log.WithError(err).Error("Error flushing NATS")
		return err
	}

	respMsg, nextErr := sync.NextMsg(1 * time.Minute)
	if nextErr != nil {
		log.WithError(nextErr).Errorf("Error reading NATS msg")
		return nextErr
	}

	var resp models.TCPConnectResponse
	jsonErr = json.Unmarshal(respMsg.Data, &resp)
	if jsonErr != nil {
		return jsonErr
	}

	var respError error
	if resp.State.IsFailed() {
		errorMsg := fmt.Sprintf("%s: %s", resp.State, resp.StateDetails)
		respError = errors.New(errorMsg)
	}

	log.WithFields(
		log.Fields{
			"status":  resp.State,
			"details": resp.StateDetails,
		},
	).Debug("End Connection request")

	return respError
}

func (s *server) connectHandlerNats(c *gin.Context) {
	clientID := FetchClientIDFromProxyAuthHeader(c)
	connectionUUID := uuid.New().String()

	if err := s.sendConnectionRequest(connectionUUID, clientID, c.Request.Host); err != nil {
		log.WithError(err).
			WithField("clientID", clientID).
			Error("Unable to handle connection request")

		c.String(http.StatusInternalServerError, "unable to make tunnel %s", err.Error())
		return
	}

	outBoundSubject := httpproxy.MakeHttpsMessageSubject(clientID, connectionUUID)
	inBoundSubject := httpproxy.MakeHttpsMessageSubject(s.locationID, connectionUUID)
	c.JSON(http.StatusOK, "")

	sourceConnection, _, err := c.Writer.Hijack()
	if err != nil {
		c.String(http.StatusInternalServerError, "unable hijack connection %s", err.Error())
		return
	}

	go func() {
		if err := models.StartBiDiNatsTunnel(
			s.natsClient,
			outBoundSubject,
			inBoundSubject,
			connectionUUID,
			sourceConnection,
		); err != nil {
			log.WithError(err).Error("failed to start bidi nats tunnel")
		}
	}()

}

func connectHandlerLocal(c *gin.Context) {
	clientID := FetchClientIDFromProxyAuthHeader(c)
	connectionUUID := uuid.New().String()
	log.Debugf("Got connect request from client %s uuid=%s", clientID, connectionUUID)
	log.Debugf("Target Host: %s", c.Request.Host)

	destConn, err := net.DialTimeout("tcp", c.Request.Host, 10*time.Second)
	if err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	c.JSON(http.StatusOK, "")
	srcCon, _, err := c.Writer.Hijack()
	if err != nil {
		log.Errorf("Unable to hijack connection %s", err.Error())
	}
	go transferTcpDataToTcp(srcCon, destConn)
	go transferTcpDataToTcp(destConn, srcCon)
}

func transferTcpDataToTcp(src io.ReadCloser, dest io.WriteCloser) {
	defer func() {
		if err := src.Close(); err != nil {
			log.WithError(err).Error("failed to close source socket")
		}
	}()

	defer func() {
		if err := dest.Close(); err != nil {
			log.WithError(err).Error("failed to close destination socket")
		}
	}()

	buf := make([]byte, 1024)
	for {
		log.Debug("Reading Data")
		readLen, readErr := src.Read(buf)
		if readErr != nil {
			log.WithError(readErr).Errorf("Error reading data")
			break
		}

		log.Debugf("Read %d bytes", readLen)
		if readLen > 0 {

			writeBuffer := buf[:readLen]
			if writeLen, writeErr := dest.Write(writeBuffer); writeErr != nil {
				log.WithError(writeErr).Errorf("Error writing data")
				break

			} else {
				log.Debugf("Wrote %d bytes", writeLen)
			}
		}
	}
	log.Debug("Terminating")
	//send terminate
}

func (s *server) RouteHandler(c *gin.Context) {

	clientID := FetchClientIDFromProxyAuthHeader(c)

	if clientID == "" {

		//normalize out the string
		tmp := c.Request.RequestURI
		if !strings.HasSuffix(tmp, "/") {
			tmp = fmt.Sprintf("%s/", tmp)
		}

		if strings.HasSuffix(tmp, "/about/") {
			aboutGetUnversioned(c)
			return
		}

		if strings.HasSuffix(tmp, "/healthcheck/") {
			healthCheckGetUnversioned(c)
			return
		}

		if strings.HasSuffix(tmp, "/metrics/") {
			metricsEndpoint(c)
			return
		}

		c.JSON(http.StatusServiceUnavailable, "")
		return
	}

	msg, err := NewHttpApiReqMessageFromHttpRequest(c.Request)
	if err != nil {
		log.WithError(err).Error("Failed to get http api request message")
		code, resp := HandleError(c, err)
		c.JSON(code, resp)
		return
	}

	jsonBits, jsonErr := json.Marshal(&msg)
	if jsonErr != nil {
		log.WithError(jsonErr).Errorf("Error marshalling http api request message")
		code, resp := HandleError(c, jsonErr)
		c.JSON(code, resp)
		return
	}

	reply := httpproxy.MakeReplyMessageSubject(s.locationID, s.test)
	sub := httpproxy.MakeMessageSubject(clientID, httpproxy.HTTP_PROXY_API_ID)
	sync, err := s.natsClient.SubscribeSync(reply)
	if err != nil {
		log.WithError(err).Errorf("Error connecting to NATS")
		code, resp := HandleError(c, err)
		c.JSON(code, resp)
		return
	}

	if err = sync.AutoUnsubscribe(1); err != nil {
		log.WithError(err).Errorf("failed to auto unsubscribe to natssync system")
	}
	if err = s.natsClient.PublishRequest(sub, reply, jsonBits); err != nil {
		log.WithError(err).Errorf("failed to publish requests")
	}
	if err = s.natsClient.Flush(); err != nil {
		log.WithError(err).Errorf("failed to flush natssync client")
	}

	respMsg, nextErr := sync.NextMsg(1 * time.Minute)
	if nextErr != nil {
		log.Errorf("Error reading nats msg %s", nextErr.Error())
		code, resp := HandleError(c, nextErr)
		c.JSON(code, resp)
		return
	}

	var k8sResp HttpApiResponseMessage
	if jsonErr := json.Unmarshal(respMsg.Data, &k8sResp); jsonErr != nil {
		log.WithError(jsonErr).Error("Error decoding NATS Message")
		code, resp := HandleError(c, jsonErr)
		c.JSON(code, resp)
		return
	}

	contentTypeHeader, gotHeader := k8sResp.Headers["Content-Type"]
	var contentType string
	if gotHeader {
		contentType = contentTypeHeader
	}

	//if the content type is JSON, give it back pretty (Maybe this slows things down.... food for though)
	if contentType == "application/json" {
		dataMap := make(map[string]interface{})
		xerr := json.Unmarshal([]byte(k8sResp.RespBody), &dataMap)
		if xerr != nil {
			log.Errorf(xerr.Error())
		}
		c.IndentedJSON(k8sResp.HttpStatusCode, dataMap)

	} else {
		//else just stream it back to them
		respLen := int64(len(k8sResp.RespBody))
		c.DataFromReader(k8sResp.HttpStatusCode, respLen, contentType, strings.NewReader(k8sResp.RespBody), k8sResp.Headers)
	}
}

type AboutResponse struct {
	AppVersion string `json:"appVersion"`
}

func aboutGetUnversioned(c *gin.Context) {
	var resp AboutResponse
	resp.AppVersion = "1.0.0"

	c.JSON(http.StatusOK, resp)
}

func healthCheckGetUnversioned(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

// TODO write this endpoint out
func metricsEndpoint(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, "")
}
