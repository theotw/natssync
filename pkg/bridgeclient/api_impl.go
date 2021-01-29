/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/msgs"

	"net/http"
)

const WAIT_MAX = 30

var registed bool

func handleGetRegister(c *gin.Context) {
	if registed {
		c.JSON(200, "")
	} else {
		c.JSON(400, "")
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
	var req v1.RegisterOnPremReq
	premID := uuid.New().String()
	log.Debugf("Generating new key for prem ID %s", premID)
	err := msgs.GenerateAndSaveKey(premID)
	if err != nil {
		log.Errorf("Error generating key %s", err.Error())
		code, ret := bridgemodel.HandleErrors(c, err)
		c.JSON(code, &ret)
		return
	}
	store := msgs.GetKeyStore()
	pkBits, err := store.ReadPublicKeyData(premID)
	if err != nil {
		log.Errorf("Error reading pub key")
		code, ret := bridgemodel.HandleErrors(c, err)
		c.JSON(code, &ret)
		return
	}
	req.PublicKey = string(pkBits)
	jsonBits, _ := json.Marshal(&req)
	serverURL := pkg.GetEnvWithDefaults("CLOUD_BRIDGE_URL", "http://localhost:8080")
	url := fmt.Sprintf("%s/event-bridge/1/register/", serverURL)
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
	c.JSON(200, "")

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
	c.Redirect(302, "/onprem-bridge/api/index_onprem_v1.html")
}

func uiGetHandler(c *gin.Context) {
	c.Redirect(302, "/onprem-bridge/ui/index.html")
}

func metricGetHandlers(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
