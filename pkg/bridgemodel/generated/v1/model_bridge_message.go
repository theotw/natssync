/*
 * On Prem cloud side REST APIBridge REST API
 *
 * Cloud side service to move messages between on prem and cloud
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package v1

type BridgeMessage struct {

	// Format version of this message.  This indicates how the message is encrypted
	FormatVersion string `json:"formatVersion,omitempty"`

	ClientID string `json:"clientID,omitempty"`

	// Encrypted message data.
	MessageData string `json:"messageData,omitempty"`
}
