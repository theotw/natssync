/*
 * On Prem client side REST API
 *
 * Client side service to move messages between on prem and cloud
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package v1

type RegistrationResponse struct {

	// the ID generated by the cloud that identifies this client
	LocationID string `json:"locationID,omitempty"`
}