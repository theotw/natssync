package msgs

import (
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
)

type CertRotationRequest struct {
	PublicKeyPackage MessageEnvelope  `json:"publicKey,omitempty"`
	PremID           string           `json:"premID,omitempty"`
	AuthChallenge    v1.AuthChallenge `json:"authChallenge,omitempty"`
}