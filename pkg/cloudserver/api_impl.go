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
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/bridgemodel/errors"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
)

func handleGetMessages(c *gin.Context) {
	clientID := c.Param("premid")
	log.Tracef("Handling get message request for clientID %s", clientID)
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

	ret := make([]v1.BridgeMessage, 0)
	metrics.IncrementTotalQueries(1)
	sub := GetSubscriptionForClient(clientID)
	if sub != nil {
		m, e := sub.NextMsg(3 * time.Second)
		if e == nil {
			plainMsg := new(bridgemodel.NatsMessage)
			plainMsg.Data = m.Data
			plainMsg.Reply = m.Reply
			plainMsg.Subject = m.Subject
			envelope, err2 := msgs.PutObjectInEnvelope(plainMsg, msgs.CLOUD_ID, clientID)
			if err2 == nil {
				jsonData, marshelError := json.Marshal(&envelope)
				if marshelError == nil {
					var bridgeMsg v1.BridgeMessage
					bridgeMsg.MessageData = string(jsonData)
					bridgeMsg.FormatVersion = "1"
					bridgeMsg.ClientID = clientID
					ret = append(ret, bridgeMsg)
				} else {
					log.Errorf("Error marshelling message in envelope %s \n", marshelError.Error())
				}
			} else {
				log.Errorf("Error putting message in envelope %s \n", err2.Error())
			}
		} else {
			log.Tracef("Error fetching messages from subscription for %s error %s", clientID, e.Error())
		}
	} else {
		//make this trace because its really just a timeout
		log.Errorf("Got a request for messages for a client ID that has no subscription %s \n", clientID)
	}

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
	nc := bridgemodel.GetNatsConnection()
	errors := make([]*v1.ErrorResponse, 0)
	for _, msg := range in.Messages {
		var envl msgs.MessageEnvelope
		err := json.Unmarshal([]byte(msg.MessageData), &envl)
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

func handlePostUnRegister(c *gin.Context) {
	var in *v1.UnRegisterOnPremReq
	response, e := sendUnRegRequestToAuthServer(c, in)
	if e != nil {
		metrics.IncrementClientUnRegistrationFailure(1)
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	} else {
		metrics.IncrementClientUnRegistrationSuccess(1)
	}
	if !response.Success {
		ierr := errors.NewInternalError(errors.BRIDGE_ERROR, errors.INVALID_REGISTRATION_REQ, nil)
		c.JSON(bridgemodel.HandleError(c, ierr))
		return
	}
	locationID := msgs.GetKeyStore().LoadLocationID()
	if locationID == "" {
		err := errors.NewInternalError(errors.BRIDGE_ERROR, errors.INVALID_LOCATION_ID, nil)
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}

	err := msgs.GetKeyStore().RemoveLocation(locationID)
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	c.JSON(201, nil)
}

func handlePostRegister(c *gin.Context) {
	var in *v1.RegisterOnPremReq
	if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
		var err error
		in, err = handleMultipartFormRegistration(c)
		if err != nil {
			metrics.IncrementClientRegistrationFailure(1)
			c.JSON(bridgemodel.HandleError(c, err))
		}
	} else {
		in = new(v1.RegisterOnPremReq)
		e := c.ShouldBindJSON(in)
		if e != nil {
			metrics.IncrementClientRegistrationFailure(1)
			code, ret := bridgemodel.HandleErrors(c, e)
			c.JSON(code, &ret)
			return
		}
	}
	pubKeyBits, decoderr := base64.StdEncoding.DecodeString(in.PublicKey)
	if decoderr != nil {
		ierr := errors.NewInternalError(errors.BRIDGE_ERROR, errors.INVALID_PUB_KEY, nil)
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
		metrics.IncrementClientRegistrationFailure(1)
		ierr := errors.NewInternalError(errors.BRIDGE_ERROR, errors.INVALID_PUB_KEY, nil)
		c.JSON(bridgemodel.HandleError(c, ierr))
		return
	}
	response, e := sendRegRequestToAuthServer(c, in)
	if e != nil {
		metrics.IncrementClientRegistrationFailure(1)
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	} else {
		metrics.IncrementClientRegistrationSuccess(1)
	}

	if !response.Success {
		ierr := errors.NewInternalError(errors.BRIDGE_ERROR, errors.INVALID_REGISTRATION_REQ, nil)
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
	nc := bridgemodel.GetNatsConnection()
	nc.Publish(bridgemodel.REGISTRATION_LIFECYCLE_ADDED, []byte(locationID))
	c.JSON(201, &resp)
}

func sendRegRequestToAuthServer(c *gin.Context, in *v1.RegisterOnPremReq) (*bridgemodel.RegistrationResponse, error) {
	timeout := time.Second * 30
	nc := bridgemodel.GetNatsConnection()
	ret := new(bridgemodel.RegistrationResponse)
	log.Tracef("Posting registratin message to nats ")
	regReq := bridgemodel.RegistrationRequest{AuthToken: in.AuthToken}
	reqBits, _ := json.Marshal(&regReq)
	respMsg, err := nc.Request(bridgemodel.REGISTRATION_AUTH_SUBJECT, reqBits, timeout)
	if err != nil {
		log.Errorf("Error sending to NATS %s", err.Error())
		return nil, err
	}
	err = json.Unmarshal(respMsg.Data, ret)
	if err != nil {
		log.Errorf("Error decoding nats response %s", err.Error())
		return nil, err
	}
	return ret, nil
}

func sendUnRegRequestToAuthServer(c *gin.Context, in *v1.UnRegisterOnPremReq) (*bridgemodel.UnRegistrationResponse, error) {
	timeout := time.Second * 30
	nc := bridgemodel.GetNatsConnection()
	ret := new(bridgemodel.UnRegistrationResponse)
	log.Tracef("Posting Unregistration message to nats ")
	unregReq := bridgemodel.UnRegistrationRequest{AuthToken: in.AuthToken}
	reqBits, _ := json.Marshal(&unregReq)
	respMsg, err := nc.Request(bridgemodel.UNREGISTRATION_AUTH_SUBJECT, reqBits, timeout)
	if err != nil {
		log.Errorf("Error sending unregister message to NATS %s", err.Error())
		return nil, err
	}
	err = json.Unmarshal(respMsg.Data, ret)
	if err != nil {
		log.Errorf("Error decoding unregister nats response %s", err.Error())
		return nil, err
	}
	return ret, nil
}

func aboutGetUnversioned(c *gin.Context) {
	var resp v1.AboutResponse
	resp.AppVersion = pkg.VERSION  // Run `make generate` to create version
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
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
