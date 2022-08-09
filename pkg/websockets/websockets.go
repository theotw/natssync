package websockets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/msgs"
)

const (
	hostAddress = "server"
)

func client() error {
	hostUrl := url.URL{
		Scheme: "https",
		Host:   hostAddress,
		Path:   "/",
	}
	conn, _, err := websocket.DefaultDialer.Dial(hostUrl.String(), nil)
	if err != nil {
		return err
	}

	msg := []byte("Hello, world")
	err = conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	_, msg, err = conn.ReadMessage()
	if err != nil {
		return err
	}

	fmt.Println(string(msg))
	return nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
}

func HandleConnectionRequest(ctx *gin.Context) {
	log.Info("Handling websocket connection request")

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.WithError(err).Error("WebSocket connection upgrade failed")
		return
	}
	log.Info("WebSocket connection started")

	clientID := ctx.Param("premid")
	subObject, exists := ctx.Get("subscription")
	sub, ok := subObject.(*nats.Subscription)
	if !exists || !ok || sub == nil {
		log.WithField("clientID", clientID).Error("No subscription for client")
	}

	go messageReceiver(conn, clientID)
	go messageSender(conn, clientID, sub)
}

func messageReceiver(conn *websocket.Conn, clientID string) {
	for {
		messageType, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.WithError(err).Error("WebSocket read message failure")
			return
		}

		log.WithField("type", messageType).WithField("data", messageBytes).Info("Received WebSocket message")

		handleReadMessage(messageBytes, clientID)
	}
}

func messageSender(conn *websocket.Conn, clientID string, sub *nats.Subscription) {
	handleGetMessages(conn, clientID, sub)
}

func handleReadMessage(messageBytes []byte, clientID string) {
	var request v1.BridgeMessagePostReq
	err := json.Unmarshal(messageBytes, &request)
	if err != nil {
		log.WithError(err).
			WithField("clientID", clientID).
			WithField("request", messageBytes).
			Error("Failure to unmarshal request")
		return
	}

	if !msgs.ValidateAuthChallenge(clientID, &request.AuthChallenge) {
		log.WithError(err).WithField("clientID", clientID).Error("Got invalid message auth request")
		return
	}

	nc := bridgemodel.GetNatsConnection()
	for _, msg := range request.Messages {
		var envl msgs.MessageEnvelope
		err = json.Unmarshal([]byte(msg.MessageData), &envl)
		if err != nil {
			log.Errorf("Error unmarshalling envelope %s", err.Error())
			continue
		}

		var natmsg bridgemodel.NatsMessage
		err = msgs.PullObjectFromEnvelope(&natmsg, &envl)
		if err != nil {
			log.Errorf("Error decoding envelope %s", err.Error())
			continue
		}
		log.Tracef("Posting message to nats sub=%s, repl=%s", natmsg.Subject, natmsg.Reply)
		if strings.HasSuffix(natmsg.Subject, msgs.ECHO_SUBJECT_BASE) {
			if len(natmsg.Reply) == 0 {
				log.Errorf("Got an echo message with no reply")
			} else {
				var echomsg nats.Msg
				echomsg.Subject = fmt.Sprintf("%s.bridge-server-post", natmsg.Reply)
				startpost := time.Now()
				tmpstring := startpost.Format("20060102-15:04:05.000")
				echoMsg := fmt.Sprintf("%s | %s", tmpstring, "message-server")
				echomsg.Data = []byte(echoMsg)
				nc.Publish(echomsg.Subject, echomsg.Data)
			}
		}
		if len(natmsg.Reply) > 0 {
			nc.PublishRequest(natmsg.Subject, natmsg.Reply, natmsg.Data)
		} else {
			nc.Publish(natmsg.Subject, natmsg.Data)
		}
		nc.Flush()
	}
}

func handleGetMessages(conn *websocket.Conn, clientID string, sub *nats.Subscription) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.WithError(err).Warning("Error attempting to close websocket connection")
		}
	}()

	timeoutStr := pkg.GetEnvWithDefaults("NATSSYNC_MSG_WAIT_TIMEOUT", "5")
	waitTimeout, numErr := strconv.ParseInt(timeoutStr, 10, 16)
	if numErr != nil {
		waitTimeout = 5
	}

	if sub == nil {
		//make this trace because its really just a timeout
		log.Errorf("Got a request for messages for a client ID that has no subscription %s \n", clientID)
		return
	}

	for {
		msg, err := sub.NextMsg(time.Duration(waitTimeout) * time.Millisecond)
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}
			log.WithError(err).Error("Failure to get message from NATS")
			return
		}

		if strings.HasSuffix(msg.Subject, msgs.ECHO_SUBJECT_BASE) {
			if len(msg.Reply) == 0 {
				log.Errorf("Got an echo message with no reply")
			} else {
				var echomsg nats.Msg
				echomsg.Subject = fmt.Sprintf("%s.bridge-server-get", msg.Reply)
				startpost := time.Now()
				tmpstring := startpost.Format("20060102-15:04:05.000")
				echoMsg := fmt.Sprintf("%s | %s", tmpstring, "message-server")
				echomsg.Data = []byte(echoMsg)
				bridgemodel.GetNatsConnection().Publish(echomsg.Subject, echomsg.Data)
			}
		}

		plainMsg := newMsgFromNatsMsg(msg)
		envelope, err := msgs.PutObjectInEnvelope(plainMsg, pkg.CLOUD_ID, clientID)
		if err != nil {
			log.WithError(err).Error("Failed to create envelope with message")
			continue
		}

		jsonData, err := json.Marshal(envelope)
		if err != nil {
			log.WithError(err).Error("Failed to marshal message in envelope")
		}

		bridgeMsg, err := newBridgeMsg(jsonData, 1, clientID)
		if err != nil {
			log.WithError(err).Error("Failed to create bridge message")
			continue
		}

		if err = conn.WriteMessage(websocket.TextMessage, bridgeMsg); err != nil {
			log.WithError(err).WithField("clientID", clientID).Error("Failed to send message to client")
		}
	}
}

func newMsgFromNatsMsg(msg *nats.Msg) *bridgemodel.NatsMessage {
	return &bridgemodel.NatsMessage{
		Data:    msg.Data,
		Reply:   msg.Reply,
		Subject: msg.Subject,
	}
}

func newBridgeMsg(data []byte, version int, clientID string) ([]byte, error) {
	msg := v1.BridgeMessage{
		MessageData:   string(data),
		FormatVersion: string(rune(version)),
		ClientID:      clientID,
	}
	return json.Marshal(msg)
}
