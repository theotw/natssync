/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package cloudserver

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/bridgemodel/errors"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/msgs"
	"math"
	"net/http"
	"time"
)

const WAIT_MAX = 30

func handleGetMessages(c *gin.Context) {
	clientID := c.Param("premid")
	fmt.Println(clientID)
	mgr := GetCacheMgr()

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
	c.JSON(200, ret)
}
func handlePostMessage(c *gin.Context) {
	clientID := c.Param("premid")
	log.Debug(clientID)
	var in v1.BridgeMessage
	e := c.ShouldBindJSON(&in)
	if e != nil {
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	}
	var envl msgs.MessageEnvelope
	err := json.Unmarshal([]byte(in.MessageData), &envl)
	if err != nil {
		log.Errorf("Error unmarshalling envelope %s", err.Error())
		code, resp := bridgemodel.HandleError(c, err)
		c.JSON(code, resp)
	}

	var natmsg bridgemodel.NatsMessage
	err = msgs.PullObjectFromEnvelope(&natmsg, &envl)
	if err != nil {
		log.Errorf("Error decoding envelope %s", err.Error())
		code, resp := bridgemodel.HandleError(c, err)
		c.JSON(code, resp)

	}
	log.Debugf("Got message %s ", natmsg.Subject)
	natsURL := pkg.GetEnvWithDefaults("NATS_SERVER_URL", "nats://127.0.0.1:4322")
	log.Infof("Connecting to NATS server %s", natsURL)
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Errorf("Error connecting to NATS %s", err.Error())
		code, resp := bridgemodel.HandleError(c, err)
		c.JSON(code, resp)

	} else {
		log.Tracef("Posting message to nats %s", natmsg.Subject)
		nc.Publish(natmsg.Subject, natmsg.Data)
		nc.Flush()
		nc.Close()
	}

}
func handlePostRegister(c *gin.Context) {
	var in v1.RegisterOnPremReq
	e := c.ShouldBindJSON(&in)
	if e != nil {
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	}
	response, e := sendRegResponse(c, in)
	if e != nil {
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	}


	if !response.Success{
		ierr:=errors.NewInernalError(errors.BRIDGE_ERROR,errors.INVALID_REGISTRATION_REQ,nil)
		c.JSON(bridgemodel.HandleError(c,ierr))
		return
	}
	locationID:=uuid.New().String()
	store := msgs.GetKeyStore()
	pubKeyBits := []byte(in.PublicKey)
	validPubKey:=false
	pubKeyBlock, _ := pem.Decode(pubKeyBits)
	if pubKeyBlock != nil{
		pubKey, perr := x509.ParsePKIXPublicKey(pubKeyBlock.Bytes)
		validPubKey=pubKey != nil && perr ==nil
	}
	if !validPubKey{
		ierr:=errors.NewInernalError(errors.BRIDGE_ERROR,errors.INVALID_REGISTRATION_REQ,nil)
		c.JSON(bridgemodel.HandleError(c,ierr))
		return
	}

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

func sendRegResponse(c *gin.Context, in v1.RegisterOnPremReq) (*bridgemodel.RegistrationResponse,error) {
	timeout := time.Second * 30

	natsURL := pkg.GetEnvWithDefaults("NATS_SERVER_URL", "nats://127.0.0.1:4322")
	log.Infof("Connecting to NATS server for regustration %s", natsURL)
	nc, err := nats.Connect(natsURL)
	ret:=new (bridgemodel.RegistrationResponse)
	if err != nil {
		log.Errorf("Error connecting to NATS %s", err.Error())
		return nil,err
	} else {
		log.Tracef("Posting message to nats ")
		regReq := bridgemodel.RegistrationRequest{UserID: in.UserID, Secret: in.Secret}
		reqBits, _ := json.Marshal(&regReq)
		respMsg, err := nc.Request(bridgemodel.REGISTRATION_AUTH_SUBJECT, reqBits, timeout)
		nc.Close()
		if err != nil{
			log.Errorf("Error sending to NATS %s", err.Error())
			return nil,err
		}
		err=json.Unmarshal(respMsg.Data,ret)
		if err != nil{
			log.Errorf("Error decoding nats response %s", err.Error())
			return nil,err
		}
	}
	return ret,nil
}
func aboutGetUnversioned(c *gin.Context) {
	var resp v1.AboutResponse
	resp.AppVersion = "1.0.0"
	resp.ApiVersions = make([]string, 0)
	resp.ApiVersions = append(resp.ApiVersions, "1")

	c.JSON(http.StatusOK, resp)
}
func healthCheckGetUnversioned(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
func swaggerUIGetHandler(c *gin.Context) {
	c.Redirect(302, "/event-bridge/api/index_v1.html")
}
