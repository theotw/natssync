/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/msgs"
	"net/url"
	"strings"
)

// WebSocketMessageHandler A web socket based implementation of the BiDiMessageHanadler
type WebSocketMessageHandler struct {
	serverURL string
}

func NewWebSocketMessageHandler(serverURL string) *WebSocketMessageHandler {
	ret := new(WebSocketMessageHandler)
	ret.serverURL=serverURL
	return ret
}
func (t *WebSocketMessageHandler) GetHandlerType() string{
	return "web-socket"
}
func (t *WebSocketMessageHandler) StartMessageHandler(clientID string) error {
	urlSplit := strings.SplitAfterN(t.serverURL, "://", 2)
	urlObject := url.URL{
		Scheme: "ws",
		Host:   urlSplit[1],
		Path:   fmt.Sprintf("/bridge-server/1/message-queue/%s/ws", clientID),
	}
	websocketURL := urlObject.String()
	log.WithField("websocketURL", websocketURL).Info("Using websocket transport")

	conn, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		log.WithError(err).WithField("url", websocketURL).Error("Failed to connect to websocket")
		return err
	}
	go t.subscribeAndSendMessageToCloud(conn,clientID)
	go t.ReadWSFromCloud(conn)
	return nil
}
func (t *WebSocketMessageHandler) StopMessageHandler() {
	//TODO Should probably do something here
}
func (t *WebSocketMessageHandler) subscribeAndSendMessageToCloud(conn *websocket.Conn, clientID string) {
	nc := bridgemodel.GetNatsConnection()
	subject := fmt.Sprintf("%s.>", msgs.NATSSYNC_MESSAGE_PREFIX)
	_, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		log.Info("Received NATS message to send to cloud via websocket")
		parsedSubject, err := msgs.ParseSubject(msg.Subject)
		if err != nil {
			log.WithError(err).Error("Failure to parse subject")
			return
		}
		if parsedSubject.LocationID == clientID {
			return
		}

		msgFormat := msgs.GetMsgFormat()
		status, err := msgFormat.ValidateMsgFormat(msg.Data, false)
		if err != nil {
			log.Errorf("Error validating the cloud event message: %s", err.Error())
			return
		}
		if !status {
			log.Errorf("Cloud event message validation failed, ignoring the message...")
			return
		}

		natmsg := bridgemodel.NatsMessage{Reply: msg.Reply, Subject: msg.Subject, Data: msg.Data}
		envelope, enverr := msgs.PutObjectInEnvelope(natmsg, clientID, pkg.CLOUD_ID)
		if enverr != nil {
			log.Errorf("Error putting msg in envelope %s", enverr.Error())
			return
		}
		jsonbits, jsonerr := json.Marshal(&envelope)
		if jsonerr != nil {
			log.Errorf("Error encoding envelope to json bits, wkipping message %s", jsonerr.Error())
			return
		}
		var bmsgs []v1.BridgeMessage
		bmsg := v1.BridgeMessage{ClientID: clientID, MessageData: string(jsonbits), FormatVersion: "1"}

		request := v1.BridgeMessagePostReq{
			AuthChallenge: *msgs.NewAuthChallengeFromStoredKey(),
			Messages:      append(bmsgs, bmsg),
		}

		msgJSON, err := json.Marshal(request)
		if err != nil {
			log.WithError(err).Error("Failed to marshal msg to JSON")
			return
		}
		if err = conn.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
			log.WithError(err).Error("Failed to send message to websocket")
			return
		}
		log.Info("Message sent to cloud via websocket")
	})
	if err != nil {
		log.WithError(err).Error("Failed to subscribe to subject")
		return
	}
}

func(t *WebSocketMessageHandler) ReadWSFromCloud(conn *websocket.Conn) {
	nc := bridgemodel.GetNatsConnection()
	defer func() { conn.Close() }()
	for {
		_, msgBytes, err := conn.ReadMessage()
		log.Info("Received message from the cloud via websocket")

		if err != nil {
			log.WithError(err).Error("Failed to read websocket message")
			continue
		}

		var bridgeMsg v1.BridgeMessage
		if err = json.Unmarshal(msgBytes, &bridgeMsg); err != nil {
			log.WithError(err).Error("Failed to unmarshal message")
			continue
		}

		var env msgs.MessageEnvelope
		if err = json.Unmarshal([]byte(bridgeMsg.MessageData), &env); err != nil {
			log.Errorf("Error unmarshalling envelope %s", err.Error())
			continue
		}

		var natmsg bridgemodel.NatsMessage
		if err = msgs.PullObjectFromEnvelope(&natmsg, &env); err != nil {
			log.WithError(err).Error("Failure pulling object from envelope")
			continue
		}

		err = nc.PublishRequest(natmsg.Subject, natmsg.Reply, natmsg.Data)
		if err != nil {
			log.WithError(err).WithField("message", natmsg).Error("Error attempting to publish message")
		}
		log.Info("Published message from websocket to NATS")
	}
}