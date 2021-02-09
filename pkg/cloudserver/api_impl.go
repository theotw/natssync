/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/bridgemodel/errors"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const WAIT_MAX = 30

func handleGetMessages(c *gin.Context) {
	clientID := c.Param("premid")
	fmt.Println(clientID)
	var in v1.AuthChallenge
	e := c.ShouldBindJSON(&in)
	if e != nil {
		_, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(400, ret)
		return
	}
	if !msgs.ValidateAuthChallenge(clientID, &in) {
		c.JSON(401, "")
		return
	}

	mgr := GetCacheMgr()
	metrics.IncrementTotalQueries(1)
	now := time.Now().Unix()
	start := now
	var messages []*CachedMsg
	for now < (start + WAIT_MAX) {
		//loop looking for messages for some amount of seconds.
		//If a message shows up.. Send it, if not, try again ever couple seconds until timeout or a message shows up.
		//KISS
		var err error
		messages, err = mgr.GetMessages(clientID)
		if err != nil {
			code, resp := bridgemodel.HandleError(c, err)
			c.JSON(code, resp)
			return
		} else {
			if len(messages) == 0 {
				time.Sleep(2 * time.Second)
				now = time.Now().Unix()
			} else {
				now = math.MaxInt64
			}

		}
	}
	ret := make([]v1.BridgeMessage, len(messages))
	for i, x := range messages {
		ret[i].ClientID = clientID
		ret[i].FormatVersion = "1"
		ret[i].MessageData = x.Data
	}
	end := time.Now().Unix()
	metrics.AddCountToWaitTimes(int(end - start))

	c.JSON(200, ret)
}
func handlePostMessage(c *gin.Context) {
	clientID := c.Param("premid")
	log.Debug(clientID)
	var in v1.BridgeMessagePostReq
	e := c.ShouldBindJSON(&in)

	if e != nil {
		_, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(400, ret)
		return
	}
	if !msgs.ValidateAuthChallenge(clientID, &in.AuthChallenge) {
		c.JSON(401, "")
		return
	}
	natsURL := pkg.Config.NatsServerUrl
	log.Infof("Connecting to NATS server %s", natsURL)
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Errorf("Error connecting to NATS %s", err.Error())
		_, resp := bridgemodel.HandleError(c, err)
		c.JSON(500, resp)
		return
	}
	defer nc.Close()
	errors := make([]*v1.ErrorResponse, 0)
	for _, msg := range in.Messages {
		var envl msgs.MessageEnvelope
		err = json.Unmarshal([]byte(msg.MessageData), &envl)
		if err != nil {
			log.Errorf("Error unmarshalling envelope %s", err.Error())
			_, resp := bridgemodel.HandleError(c, err)
			errors = append(errors, resp)
			continue
		}

		var natmsg bridgemodel.NatsMessage
		err = msgs.PullObjectFromEnvelope(&natmsg, &envl)
		if err != nil {
			log.Errorf("Error decoding envelope %s", err.Error())
			_, resp := bridgemodel.HandleError(c, err)
			errors = append(errors, resp)
		}
		log.Tracef("Posting message to nats %s", natmsg.Subject)
		nc.Publish(natmsg.Subject, natmsg.Data)
		nc.Flush()
	}
	if len(errors) > 1 {
		c.JSON(400, errors)
	} else {
		c.JSON(202, "")
	}

}
func handleMultipartFormRegistration(c *gin.Context) (ret *v1.RegisterOnPremReq, reterr error) {
	contentType, params, parseErr := mime.ParseMediaType(c.Request.Header.Get("Content-Type"))
	fmt.Println(contentType)
	fmt.Println(params)
	if parseErr != nil {
		fmt.Println(parseErr.Error())
	}
	multipartReader := multipart.NewReader(c.Request.Body, params["boundary"])
	defer c.Request.Body.Close()
	ret = new(v1.RegisterOnPremReq)
	for {
		part, parseErr := multipartReader.NextPart()
		if parseErr != nil {
			if parseErr == io.EOF {
				break
			} else {
				reterr = parseErr
				return
			}
		}

		bits, err := ioutil.ReadAll(part)
		if err != nil {
			reterr = err
			return
		}
		fieldName := part.FormName()
		switch fieldName {
		case "authToken":
			{
				ret.AuthToken = string(bits)
				break
			}
		case "metaData":
			{
				ret.MetaData = string(bits)
				break
			}
		case "publicKey":
			{
				ret.PublicKey = string(bits)
				break
			}
		}
	}

	return
}
func handlePostRegister(c *gin.Context) {
	var in *v1.RegisterOnPremReq
	if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
		var err error
		in, err = handleMultipartFormRegistration(c)
		if err != nil {
			c.JSON(bridgemodel.HandleError(c, err))
		}
	} else {
		in = new(v1.RegisterOnPremReq)
		e := c.ShouldBindJSON(in)
		if e != nil {
			code, ret := bridgemodel.HandleErrors(c, e)
			c.JSON(code, &ret)
			return
		}
	}
	pubKeyBits, decoderr := base64.StdEncoding.DecodeString(in.PublicKey)
	if decoderr != nil {
		ierr := errors.NewInernalError(errors.BRIDGE_ERROR, errors.INVALID_PUB_KEY, nil)
		c.JSON(bridgemodel.HandleError(c, ierr))
		return
	}
	validPubKey := false
	pubKeyBlock, _ := pem.Decode(pubKeyBits)
	if pubKeyBlock != nil {
		pubKey, perr := x509.ParsePKIXPublicKey(pubKeyBlock.Bytes)
		validPubKey = pubKey != nil && perr == nil
	}
	if !validPubKey {
		ierr := errors.NewInernalError(errors.BRIDGE_ERROR, errors.INVALID_PUB_KEY, nil)
		c.JSON(bridgemodel.HandleError(c, ierr))
		return
	}
	response, e := sendRegRequestToAuthServer(c, in)
	if e != nil {
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	}

	if !response.Success {
		ierr := errors.NewInernalError(errors.BRIDGE_ERROR, errors.INVALID_REGISTRATION_REQ, nil)
		c.JSON(bridgemodel.HandleError(c, ierr))
		return
	}
	locationID := bridgemodel.GenerateUUID()
	store := msgs.GetKeyStore()

	var resp v1.RegisterOnPremResponse
	pkBits, err := store.ReadPublicKeyData(msgs.CLOUD_ID)
	if err != nil {
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	}
	store.WritePublicKey(locationID, pubKeyBits)
	resp.CloudPublicKey = string(pkBits)
	resp.PermId = locationID
	c.JSON(201, &resp)
}

func sendRegRequestToAuthServer(c *gin.Context, in *v1.RegisterOnPremReq) (*bridgemodel.RegistrationResponse, error) {
	timeout := time.Second * 30

	natsURL := pkg.Config.NatsServerUrl
	log.Infof("Connecting to NATS server for regustration %s", natsURL)
	nc, err := nats.Connect(natsURL)
	ret := new(bridgemodel.RegistrationResponse)
	if err != nil {
		log.Errorf("Error connecting to NATS %s", err.Error())
		return nil, err
	} else {
		log.Tracef("Posting message to nats ")
		regReq := bridgemodel.RegistrationRequest{AuthToken: in.AuthToken}
		reqBits, _ := json.Marshal(&regReq)
		respMsg, err := nc.Request(bridgemodel.REGISTRATION_AUTH_SUBJECT, reqBits, timeout)
		nc.Close()
		if err != nil {
			log.Errorf("Error sending to NATS %s", err.Error())
			return nil, err
		}
		err = json.Unmarshal(respMsg.Data, ret)
		if err != nil {
			log.Errorf("Error decoding nats response %s", err.Error())
			return nil, err
		}
	}
	return ret, nil
}
func aboutGetUnversioned(c *gin.Context) {
	var resp v1.AboutResponse
	resp.AppVersion = pkg.VERSION
	resp.ApiVersions = make([]string, 0)
	resp.ApiVersions = append(resp.ApiVersions, "1")

	c.JSON(http.StatusOK, resp)
}
func healthCheckGetUnversioned(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
func swaggerUIGetHandler(c *gin.Context) {
	c.Redirect(302, "/bridge-server/api/index_bridge_server_v1.html")
}

func metricGetHandlers(c *gin.Context) {
	depths := GetCacheMgr().GetQueueDepths()
	var total int
	for _, count := range depths {
		total = total + count
	}
	metrics.SetTotalMessagesQueued(total)
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
