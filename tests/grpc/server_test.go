/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package grpc

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg/pbgen"
	grpc "google.golang.org/grpc"
	"net"
	"testing"
)

type MessageServerTestImpl struct {
	pbgen.UnimplementedMessageServiceServer
}

func (t *MessageServerTestImpl) GetMessages(in *pbgen.RequestMessagesIn, x pbgen.MessageService_GetMessagesServer) error {
	for i := 0; i < 10; i++ {
		ret := new(pbgen.BridgeMessage)
		ret.MessageData = fmt.Sprintf("data %d", i)
		ret.ClientID = "1"
		ret.FormatVersion = "-1"
		x.Send(ret)
	}

	return nil
}
func (t *MessageServerTestImpl) PushMessage(xtc context.Context, msgIn *pbgen.PushMessageIn) (*pbgen.PushMessageOut, error) {
	ret:=new (pbgen.PushMessageOut)
	fmt.Println(msgIn.Msg.MessageData)
	return ret,nil
}



func TestGprcServer(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Fatal(fmt.Errorf("localAddresses: %v\n", err.Error()))
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			log.Printf("%v - %v,%s\n", i.Name, a.Network(),a.String())
		}
	}
	impl := new(MessageServerTestImpl)
	lis, err := net.Listen("tcp", fmt.Sprintf(":8084"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pbgen.RegisterMessageServiceServer(grpcServer, impl)
	log.Infof("Wainting 1")
	grpcServer.Serve(lis)
	log.Infof("Wainting 2")
}
