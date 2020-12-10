package cloudclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"onprembridge/pkg"
	"onprembridge/pkg/bridgemodel"
	serverv1 "onprembridge/pkg/bridgemodel/generated/cloudserver/v1"
	v1 "onprembridge/pkg/bridgemodel/generated/onpremserver/v1"
	"onprembridge/pkg/msgs"
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
	var in v1.RegisterReq
	e := c.ShouldBindJSON(&in)
	if e != nil {
		code, ret := bridgemodel.HandleErrors(c, e)
		c.JSON(code, &ret)
		return
	}
	var req serverv1.RegisterOnPremReq
	premID := uuid.New().String()
	req.PremID = premID
	log.Debugf("Generating new key for prem ID %s", premID)
	err := msgs.GenerateNewKeyPairPOCVersion(premID)
	if err != nil {
		log.Errorf("Error generating key %s", err.Error())
		code, ret := bridgemodel.HandleErrors(c, err)
		c.JSON(code, &ret)
		return
	}

	pkBits, err := msgs.ReadPublicKeyFile(premID)
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
