/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	models "github.com/theotw/natssync/pkg/k8srelay/model"
	"github.com/theotw/natssync/pkg/msgs"
	"github.com/theotw/natssync/pkg/natsmodel"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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

func (s *server) RouteHandler(c *gin.Context) {
	//TODO THIS IS WEIRD NEED TO FIGURE OUT IF WE NEED THIS
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
			metricsHandler(c)
			return
		}

		c.JSON(http.StatusServiceUnavailable, "")
		return
	}

}

type AboutResponse struct {
	AppVersion string `json:"appVersion"`
}

func aboutGetUnversioned(c *gin.Context) {
	var resp AboutResponse
	resp.AppVersion = "1.0.0"

	c.JSON(http.StatusOK, &resp)
}

func healthCheckGetUnversioned(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func metricsHandler(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

const bearer = "Bearer "

func genericHandlerHandler(c *gin.Context) {
	parse, err := url.Parse(c.Request.RequestURI)
	var userTokenWhichBecomesRouteID string
	token := c.Request.Header.Get("Authorization")
	if strings.HasPrefix(token, bearer) {
		userTokenWhichBecomesRouteID = token[len(bearer):]
	} else {
		userTokenWhichBecomesRouteID = "dev"
	}
	log.Infof(userTokenWhichBecomesRouteID)
	log.Infof("URI %s", c.Request.URL.String())
	if err != nil {
		panic(err)
	}

	var req models.CallRequest
	req.Path = parse.Path
	req.Method = c.Request.Method
	bodyBits, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithError(err).Errorf("Unable to read body on call %s - %s error: %s", c.Request.Method, parse.Path, err.Error())
	}
	req.InBody = bodyBits
	nc := natsmodel.GetNatsConnection()
	replySub := msgs.MakeNBReplySubject()
	sbMsgSub := msgs.MakeMessageSubject(userTokenWhichBecomesRouteID, models.K8SRelayRequestMessageSubjectSuffix)
	bits, err := json.Marshal(&req)
	if err != nil {
		c.Status(502)
		c.Header("Content-Type", "text/plain")
		c.Writer.Write([]byte(fmt.Sprintf(" gate way error %s", err.Error())))
		return
	}
	nm := nats.NewMsg(sbMsgSub)
	nm.Data = bits
	nm.Reply = replySub
	sync, err := nc.SubscribeSync(replySub)
	if err != nil {
		c.Status(502)
		c.Header("Content-Type", "text/plain")
		c.Writer.Write([]byte(fmt.Sprintf(" gate way error %s", err.Error())))
		return
	}
	nc.PublishMsg(nm)
	msg, err := sync.NextMsg(time.Minute * 2)
	if err != nil {
		c.Status(502)
		c.Header("Content-Type", "text/plain")
		c.Writer.Write([]byte(fmt.Sprintf(" gate way error %s", err.Error())))
		return
	}
	var respMsg models.CallResponse
	err = json.Unmarshal(msg.Data, &respMsg)
	if err != nil {
		c.Status(502)
		c.Header("Content-Type", "text/plain")
		c.Writer.Write([]byte(fmt.Sprintf(" gate way error %s", err.Error())))
		return
	}

	log.Infof("Got resp status %d ", respMsg.StatusCode)
	for k, v := range respMsg.Headers {
		log.Infof("%s = %s ", k, v[0])
		c.Header(k, v)
	}
	c.Status(respMsg.StatusCode)
	if respMsg.OutBody != nil {
		c.Writer.Write(respMsg.OutBody)
	}
	c.Writer.Flush()

}
