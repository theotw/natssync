/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/theotw/natssync/pkg/httpsproxy"
	models2 "github.com/theotw/natssync/pkg/httpsproxy/models"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type JsonError struct {
	Message string `json:"message"`
}

func HandleError(_ *gin.Context, err error) (int, interface{}) {
	ret := JsonError{Message: err.Error()}
	return 500, &ret
}
func connectHandler(c *gin.Context) {
	connectHandlerNats(c)
	//connectHandlerLocal(c)
}
func connectHandlerNats(c *gin.Context) {
	clientID := FetchClientIDFromProxyAuthHeader(c)
	connectionUUID := uuid.New().String()

	err := models2.SendConnectionRequest(connectionUUID, clientID, c.Request.Host)
	if err != nil {
		log.Errorf("Unable to handle connection request for clientID %s, error: %s", clientID, err.Error())
		c.String(500, "unable to make tunnel %s", err.Error())
		return
	}
	outBoundSubject := httpproxy.MakeHttpsMessageSubject( clientID, connectionUUID)
	locationID := httpproxy.GetMyLocationID()
	inBoundSubject := httpproxy.MakeHttpsMessageSubject(locationID, connectionUUID)
	c.JSON(200, "")
	srcCon, _, err := c.Writer.Hijack()
	go models2.StartBiDiNatsTunnel(outBoundSubject, inBoundSubject, connectionUUID, srcCon)

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

	//c.Status(200)
	c.JSON(200, "")
	srcCon, _, err := c.Writer.Hijack()
	if err != nil {
		log.Errorf("Unable to hijack connection %s", err.Error())
	}
	go transferTcpDataToTcp(srcCon, destConn)
	go transferTcpDataToTcp(destConn, srcCon)
}

func transferTcpDataToTcp(src io.ReadCloser, dest io.WriteCloser) {
	defer src.Close()
	defer dest.Close()
	buf := make([]byte, 1024)
	for {
		log.Debug("Reading Data")
		len, readErr := src.Read(buf)
		log.Debugf("Read %d bytes ", len)
		if len > 0 {
			writebuf := buf[:len]
			writeLen, writeErr := dest.Write(writebuf)
			if writeErr != nil {
				log.Errorf("Error writing data %s", writeErr.Error())
				break
			}
			log.Debugf("Wrote %d bytes ", writeLen)
		}
		if readErr != nil {
			log.Errorf("Error reading data %s", readErr.Error())
			break
		}
	}
	log.Debug("Terminating")
	//send terminate
}
func routeHandler(c *gin.Context) {
	nc := models2.GetNatsClient()
	clientID := FetchClientIDFromProxyAuthHeader(c)
	if len(clientID) == 0 {
		//normalize out the string
		tmp := c.Request.RequestURI
		if !strings.HasSuffix(tmp, "/") {
			tmp = fmt.Sprintf("%s/", tmp)
		}
		if strings.HasSuffix(c.Request.RequestURI, "/about/") {
			aboutGetUnversioned(c)
			return
		} else if strings.HasSuffix(c.Request.RequestURI, "/healthcheck/") {
			healthCheckGetUnversioned(c)
			return
		} else {
			c.JSON(503, "")
		}
	}

	var msg HttpApiReqMessage
	msg.Target = c.Request.Host
	msg.HttpMethod = c.Request.Method
	bodyBits, bodyErr := ioutil.ReadAll(c.Request.Body)
	if bodyErr != nil {
		log.Errorf("Error reading body %s", bodyErr.Error())
		code, resp := HandleError(c, bodyErr)
		c.JSON(code, resp)
		return
	}
	msg.Body = bodyBits
	msg.HttpPath = c.Request.URL.Path
	msg.Headers = make([]HttpReqHeader, 0)
	for k, v := range c.Request.Header {
		x := HttpReqHeader{Key: k, Values: v}
		msg.Headers = append(msg.Headers, x)
	}
	jsonBits, jsonErr := json.Marshal(&msg)
	if jsonErr != nil {
		log.Errorf("Error reading body %s", jsonErr.Error())
		code, resp := HandleError(c, jsonErr)
		c.JSON(code, resp)
		return
	}

	reply := httpproxy.MakeReplyMessageSubject()
	sub := httpproxy.MakeMessageSubject(clientID, httpproxy.HTTP_PROXY_API_ID)
	sync, err := nc.SubscribeSync(reply)
	if err != nil {
		log.Errorf("Error connecting to NATS %s", err.Error())
		code, resp := HandleError(c, err)
		c.JSON(code, resp)
		return
	}
	_ = sync.AutoUnsubscribe(1)
	_ = nc.PublishRequest(sub, reply, jsonBits)
	_ = nc.Flush()

	respmsg, nextErr := sync.NextMsg(1 * time.Minute)
	if nextErr != nil {
		log.Errorf("Error reading nats msg %s", nextErr.Error())
		code, resp := HandleError(c, nextErr)
		c.JSON(code, resp)
		return
	}

	var k8sResp HttpApiResponseMessage
	jsonerr := json.Unmarshal(respmsg.Data, &k8sResp)
	if jsonerr != nil {
		log.Errorf("Error decoding NATS Messafe %s", jsonerr.Error())
		code, resp := HandleError(c, jsonerr)
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
