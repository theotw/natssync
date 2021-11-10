package unittestresources

import "github.com/theotw/natssync/pkg/httpsproxy/nats"

type mockRequestHandler struct {
	httpHandlerInvoked  bool
	httpsHandlerInvoked bool
	locationID string
}

func (mrh *mockRequestHandler) HttpHandler(msg *nats.Msg) {
	mrh.httpHandlerInvoked = true
}

func (mrh *mockRequestHandler) HttpsHandler(msg *nats.Msg) {
	mrh.httpsHandlerInvoked = true
}

func (mrh *mockRequestHandler) SetLocationID(locationID string) {
	mrh.locationID = locationID
}


// =====================================
func (mrh *mockRequestHandler) InvokedHttpHandler() bool {
	return mrh.httpHandlerInvoked
}

func (mrh *mockRequestHandler) InvokedHttpsHandler() bool {
	return mrh.httpsHandlerInvoked
}

func (mrh *mockRequestHandler) GetLocationID() string {
	return mrh.locationID
}

func NewMockRequestHandler() *mockRequestHandler {
	return &mockRequestHandler{}
}
