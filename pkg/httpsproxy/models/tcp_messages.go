package models

type TCPConnectRequest struct {
	// host:port
	Destination     string `json:"destination"`
	ProxyLocationID string `json:"proxyLocationID"`
	ConnectionID    string `json:"connectionID,omitempty"`
}

type TCPConnectResponse struct {
	State        string `json:"state"` //ok means connected with blank state details
	StateDetails string `json:"stateDetails,omitempty"`
	ConnectionID string `json:"connectionID,omitempty"`
}

type TCPData struct {
	DataB64      string `json:"dataB64"`
	DataLength   int    `json:"dataLength"` //if data len is 0, then this is a close message
	SequenceID   int    `json:"sequenceID"`
	ConnectionID string `json:"connectionID"`
}

type TCPCloseRequest struct {
	ConnectionID string `json:"connectionID"`
}
