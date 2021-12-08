package models

type TCPConnectRequest struct {
	// host:port
	Destination     string `json:"destination"`
	ProxyLocationID string `json:"proxyLocationID"`
	ConnectionID    string `json:"connectionID,omitempty"`
}

type TCPConnectState string

const (
	TCPConnectStateOK     TCPConnectState = "ok"
	TCPConnectStateFailed TCPConnectState = "failed"
)

func (state TCPConnectState) IsOk() bool {
	return state == TCPConnectStateOK
}

func (state TCPConnectState) IsFailed() bool {
	return state == TCPConnectStateFailed
}

type TCPConnectResponse struct {
	State        TCPConnectState `json:"state"` //ok means connected with blank state details
	StateDetails string          `json:"stateDetails,omitempty"`
	ConnectionID string          `json:"connectionID,omitempty"`
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
