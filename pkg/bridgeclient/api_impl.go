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
func handlePostRegister(c *gin.Context) {
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
	resp, err := http.DefaultClient.Post(url, "application/json", bytes.NewReader(jsonBits))

	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	if resp.StatusCode >= 300 {
		code, response := bridgemodel.HandleError(c, errors.New("invalid status "+resp.Status))
		c.JSON(code, response)
		return
	}
	bits, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		code, response := bridgemodel.HandleError(c, err)
		c.JSON(code, response)
		return
	}
	var regResp serverv1.RegisterOnPremResponse
	json.Unmarshal(bits, &regResp)
	msgs.SaveKeyPair(regResp.PermId, pair)
	msgs.GetKeyStore().WritePublicKey(msgs.CLOUD_ID, []byte(regResp.CloudPublicKey))
	//this step must be last, other parts of the code watch for this key
	msgs.GetKeyStore().SaveLocationID(regResp.PermId)

	c.JSON(200, "")

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
	c.Redirect(302, "/bridge-client/api/index_bridge_client_v1.html")
}
