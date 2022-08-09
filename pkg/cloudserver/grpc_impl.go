/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
	"github.com/theotw/natssync/pkg/pbgen"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
	"time"
)

type MessageServerImpl struct {
	pbgen.UnimplementedMessageServiceServer
}

func (t *MessageServerImpl) RunServer(port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pbgen.RegisterMessageServiceServer(grpcServer, t)
	go func() {
		log.Infof("Starting Server")
		grpcServer.Serve(lis)
		log.Infof("Server finished")
	}()
	return nil
}
func (t *MessageServerImpl) GetMessages(in *pbgen.RequestMessagesIn, x pbgen.MessageService_GetMessagesServer) error {

	timeoutStr := pkg.GetEnvWithDefaults("NATSSYNC_MSG_WAIT_TIMEOUT", "5")
	maxMsgHoldStr := pkg.GetEnvWithDefaults("NATSSYNC__MAX_MSG_HOLD", "512")
	waitTimeout, numErr := strconv.ParseInt(timeoutStr, 10, 16)
	if numErr != nil {
		waitTimeout = 5
	}
	maxQueueSize, numErr := strconv.ParseInt(maxMsgHoldStr, 10, 16)
	if numErr != nil {
		waitTimeout = 512
	}

	clientID := in.ClientID
	log.Tracef("Handling get message request for clientID %s", clientID)
	var authData v1.AuthChallenge
	authData.AuthChallengeA = in.Auth.AuthChallengeA
	authData.AuthChellengeB = in.Auth.AuthChallengeB
	if !msgs.ValidateAuthChallenge(clientID, &authData) {
		errors.New("invalid auth")
	}

	ret := make([]v1.BridgeMessage, 0)
	metrics.IncrementTotalQueries(1)
	sub := GetSubscriptionForClient(clientID)
	start := time.Now()
	if sub != nil {
		keepWaiting := true
		for keepWaiting {
			m, e := sub.NextMsg(time.Duration(waitTimeout) * time.Millisecond)
			if e == nil {
				if strings.HasSuffix(m.Subject, msgs.ECHO_SUBJECT_BASE) {
					if len(m.Reply) == 0 {
						log.Errorf("Got an echo message with no reply")
					} else {
						var echomsg nats.Msg
						echomsg.Subject = fmt.Sprintf("%s.bridge-server-get", m.Reply)
						startpost := time.Now()
						tmpstring := startpost.Format("20060102-15:04:05.000")
						echoMsg := fmt.Sprintf("%s | %s", tmpstring, "message-server")
						echomsg.Data = []byte(echoMsg)
						bridgemodel.GetNatsConnection().Publish(echomsg.Subject, echomsg.Data)
					}
				}
				plainMsg := new(bridgemodel.NatsMessage)
				plainMsg.Data = m.Data
				plainMsg.Reply = m.Reply
				plainMsg.Subject = m.Subject

				var envelopErr error
				var envelope *msgs.MessageEnvelope
				envelope, envelopErr = msgs.PutObjectInEnvelope(plainMsg, pkg.CLOUD_ID, clientID)

				if envelopErr == nil {
					jsonData, marshelError := json.Marshal(&envelope)
					if marshelError == nil {
						var bridgeMsg pbgen.BridgeMessage
						bridgeMsg.MessageData = string(jsonData)
						bridgeMsg.FormatVersion = "1"
						bridgeMsg.ClientID = clientID
						x.Send(&bridgeMsg)
					} else {
						log.Errorf("Error marshelling message in envelope %s \n", marshelError.Error())
					}
				} else {
					log.Errorf("Error putting message in envelope %s \n", envelopErr.Error())
				}
				keepWaiting = len(ret) < int(maxQueueSize)
			} else {
				keepWaiting = false
				t := time.Now()
				keepWaiting = t.Sub(start) < 30*time.Second && len(ret) == 0
			}
		}
	} else {
		//make this trace because its really just a timeout
		log.Errorf("Got a request for messages for a client ID that has no subscription %s \n", clientID)
	}
	return nil
}
func (t *MessageServerImpl) PushMessage(xtc context.Context, msgIn *pbgen.PushMessageIn) (*pbgen.PushMessageOut, error) {
	clientID := msgIn.Msg.ClientID
	ret := new(pbgen.PushMessageOut)
	log.Debug(clientID)
	//var in v1.BridgeMessagePostReq
	var auth v1.AuthChallenge
	auth.AuthChallengeA = msgIn.Auth.AuthChallengeA
	auth.AuthChellengeB = msgIn.Auth.AuthChallengeB
	log.Debugf("PushMessage: in: %v", msgIn)
	if !msgs.ValidateAuthChallenge(clientID, &auth) {
		log.Errorf("Got invalid message auth request in post messages %s", clientID)
		return nil, errors.New("auth error")
	}
	nc := bridgemodel.GetNatsConnection()
	var envl msgs.MessageEnvelope
	err := json.Unmarshal([]byte(msgIn.Msg.MessageData), &envl)
	if err != nil {
		log.Errorf("Error unmarshalling envelope %s", err.Error())
		return ret, err
	}

	var natmsg bridgemodel.NatsMessage
	err = msgs.PullObjectFromEnvelope(&natmsg, &envl)
	if err != nil {
		log.Errorf("Error decoding envelope %s", err.Error())
		return ret, err
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
		log.Debugf("publishing data %v to subject %s", natmsg.Data, natmsg.Subject)
		nc.Publish(natmsg.Subject+"local", natmsg.Data)
	}
	nc.Flush()
	return ret, nil
}

func NewGRPCMessageServerImpl() *MessageServerImpl {
	ret := new(MessageServerImpl)
	return ret
}
