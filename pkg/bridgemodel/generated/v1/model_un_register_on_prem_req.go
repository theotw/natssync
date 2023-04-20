/*
 * On Prem cloud side REST APIBridge REST API
 *
 * Cloud side service to move messages between on prem and cloud
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package v1

type UnRegisterOnPremReq struct {

	// the format of the auth token can be what ever is needed for authentication.  The auth service used on the back end will determine that.  In its simpliest form it is userID:Password
	AuthToken string `json:"authToken,omitempty"`

	// the meta data used to visually identify the user/on prem instance.  this string must be unique to the server but can be anything and be changed
	MetaData string `json:"metaData,omitempty"`
}
