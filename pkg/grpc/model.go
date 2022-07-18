/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package grpc

import (
	"context"
	"fmt"
	"github.com/theotw/natssync/pkg/pbgen"
)

type MessageServerImpl struct {
	pbgen.UnimplementedMessageServiceServer
}


func (t *MessageServerImpl) GetMessage(context.Context, *pbgen.RequestMessagesIn) (*pbgen.BridgeMessage, error){
	ret:=new (pbgen.BridgeMessage)
	ret.MessageData="data"
	ret.ClientID="1"
	ret.FormatVersion="-1"
	return ret,nil
}
func (t *MessageServerImpl) GetMessages(in *pbgen.RequestMessagesIn,x pbgen.MessageService_GetMessagesServer) error{
	for i:=0;i<10;i++ {
		ret := new(pbgen.BridgeMessage)
		ret.MessageData = fmt.Sprintf("data %d",i)
		ret.ClientID = "1"
		ret.FormatVersion = "-1"
		x.Send(ret)
	}
	return nil
}
func (t *MessageServerImpl) PushMessage(context.Context, *pbgen.BridgeMessage) (*pbgen.PushMessageOut, error){
	ret:=new (pbgen.PushMessageOut)
	return ret, nil
}

func NewMessageServerImpl() pbgen.MessageServiceServer{
	ret:=new (MessageServerImpl)
	return ret
}