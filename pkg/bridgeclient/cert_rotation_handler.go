package cloudclient

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/msgs"
	"github.com/theotw/natssync/pkg/persistence"
)

const (
	certRotationUrlFormat = "%v/bridge-server/1/register-certificate"
	connectionTimeout     = 30 * time.Second
)

type certRotationHandler struct {
	clientID        string
	client          *http.Client
	store           persistence.LocationKeyStore
	certRotationUrl string
}

func NewCertRotationHandler(cloudServerUri string, clientID string) *certRotationHandler {
	client := http.DefaultClient
	client.Timeout = connectionTimeout

	return NewCertRotationHandlerDetailed(cloudServerUri, clientID, client)
}

func NewCertRotationHandlerDetailed(
	cloudServerUri string,
	clientID string,
	client *http.Client,
) *certRotationHandler {

	return &certRotationHandler{
		client:          client,
		clientID:        clientID,
		store:           persistence.GetKeyStore(),
		certRotationUrl: fmt.Sprintf(certRotationUrlFormat, cloudServerUri),
	}
}

func (crh *certRotationHandler) HandleCertRotation() error {
	log.Infof("Handling cert rotation")

	payload := new(msgs.CertRotationRequest)
	pair, err := msgs.GenerateNewKeyPair()
	if err != nil {
		log.WithError(err).Errorf("Error generating keys")
		return err
	}

	pubKeyBits, err := x509.MarshalPKIXPublicKey(&pair.PublicKey)
	if err != nil {
		log.WithError(err).Errorf("Failed to marshall public key")
		return err
	}

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBits,
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, publicKeyBlock)
	if err != nil {
		log.WithError(err).Errorf("Failed to encode public key")
		return err
	}

	envelope, enverr := msgs.PutMessageInEnvelopeV3(buf.Bytes(), crh.clientID, pkg.CLOUD_ID)
	if enverr != nil {
		return err
	}

	payload.PremID = crh.store.LoadLocationID("")
	payload.PublicKeyPackage = *envelope
	payload.AuthChallenge = *msgs.NewAuthChallenge("")

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	response, err := crh.client.Post(crh.certRotationUrl, "application/json", bytes.NewReader(payloadBytes))
	if err != nil || response.StatusCode != http.StatusNoContent {
		statusCode := 0
		if response != nil {
			statusCode = response.StatusCode
		}

		log.WithError(err).
			WithField("statusCode", statusCode).
			Errorf("Failed to rotate certificates")
		return fmt.Errorf("failed to rotate certificates")
	}

	err = msgs.SaveKeyPair(payload.PremID, pair)
	if err != nil {
		log.WithError(err).WithField("PremID", payload.PremID).Errorf("failed to save keypair")
		return err
	}

	return nil
}
