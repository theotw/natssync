/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	v1 "github.com/theotw/natssync/pkg/bridgeclient/generated/v1"
	"github.com/theotw/natssync/pkg/bridgemodel"
	serverv1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/msgs"
	"io/ioutil"

	"net/http"
)

const WAIT_MAX = 30

func handleGetRegister(c *gin.Context) {
	store := msgs.GetKeyStore()
	locationID := store.LoadLocationID()
	if len(locationID) > 0 {
		c.JSON(200, "")
	} else {
		c.JSON(400, "")
	}
}

func handlePostUnRegister(c *gin.Context) {
	log.Debug("Handling unregistration post request")
	var in v1.UnRegisterReq
	e := c.ShouldBindJSON(&in)
	if e != nil {
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	}

	locationID := msgs.GetKeyStore().LoadLocationID()
	if locationID == "" {
		err := errors.New("Failed to load locationID")
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}

	// Call Unregister in the server
	var req serverv1.UnRegisterOnPremReq
	req.AuthToken = in.AuthToken
	req.MetaData = msgs.GetKeyStore().LoadLocationID()
	jsonBits, _ := json.Marshal(&req)
	url := fmt.Sprintf("%s/bridge-server/1/unregister/", pkg.Config.CloudBridgeUrl)

	log.Infof("Calling Unregister with cloud server %s for location %s", url, locationID)
	resp, err := http.DefaultClient.Post(url, "application/json", bytes.NewReader(jsonBits))
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	log.Debugf("Unregistration response status code %d", resp.StatusCode)
	if resp.StatusCode >= 300 {
		code, response := bridgemodel.HandleError(c, errors.New("invalid status "+resp.Status))
		c.JSON(code, response)
		return
	}

	err = msgs.GetKeyStore().RemoveLocation(locationID)
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	err = msgs.GetKeyStore().SaveLocationID("")
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	c.JSON(201, nil)
}

func handlePostRegister(c *gin.Context) {
	log.Debug("Handling registration post request")
	var in v1.RegisterReq
	e := c.ShouldBindJSON(&in)
	if e != nil {
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	}
	var req serverv1.RegisterOnPremReq
	pair, err := msgs.GenerateNewKeyPair()
	if err != nil {
		log.Errorf("Error generating key %s", err.Error())
		code, ret := bridgemodel.HandleErrors(c, err)
		c.JSON(code, &ret)
		return
	}
	pubKeyBits, err := x509.MarshalPKIXPublicKey(&pair.PublicKey)
	if err != nil {
		code, ret := bridgemodel.HandleErrors(c, err)
		c.JSON(code, &ret)
		return
	}

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBits,
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, publicKeyBlock)
	if err != nil {
		code, ret := bridgemodel.HandleErrors(c, err)
		c.JSON(code, &ret)
		return
	}

	req.PublicKey = base64.StdEncoding.EncodeToString(buf.Bytes())
	req.AuthToken = in.AuthToken
	req.MetaData = in.MetaData
	jsonBits, _ := json.Marshal(&req)
	url := fmt.Sprintf("%s/bridge-server/1/register/", pkg.Config.CloudBridgeUrl)

	log.Infof("Registering with cloud server %s", url)
	resp, err := http.DefaultClient.Post(url, "application/json", bytes.NewReader(jsonBits))

	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}

	log.Debugf("Registration response status code %d", resp.StatusCode)
	if resp.StatusCode >= 300 {
		code, response := bridgemodel.HandleError(c, errors.New("invalid status "+resp.Status))
		c.JSON(code, response)
		return
	}
	log.Debugf("Registration response body %s", resp.Body)
	bits, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	var regResp serverv1.RegisterOnPremResponse
	err = json.Unmarshal(bits, &regResp)
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	err = msgs.SaveKeyPair(regResp.PermId, pair)
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	err = msgs.GetKeyStore().WritePublicKey(msgs.CLOUD_ID, []byte(regResp.CloudPublicKey))
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	//this step must be last, other parts of the code watch for this key
	err = msgs.GetKeyStore().SaveLocationID(regResp.PermId)
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	ret := new(v1.RegistrationResponse)
	ret.LocationID = regResp.PermId
	c.JSON(200, ret)
}

func registrationGetHandler(c *gin.Context) {
	ret := new(v1.RegistrationResponse)
	ret.LocationID = msgs.GetKeyStore().LoadLocationID()
	c.JSON(200, ret)
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
	c.Redirect(302, "/bridge-client/api/index_bridge_client_v1.html")
}

func metricGetHandlers(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
