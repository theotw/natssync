/*
 * On Prem client side REST API
 *
 * Client side service to move messages between on prem and cloud
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package v1

type UnRegisterReq struct {

	// Auth token for the registration request
	AuthToken string `json:"authToken,omitempty"`
}