/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package grpc

import (
	"context"
	"fmt"
	"github.com/theotw/natssync/pkg/pbgen"
	grpc "google.golang.org/grpc"
	"testing"
)

func TestGRPCClient(t *testing.T){
	conn, err := grpc.Dial("localhost:8082",grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Unable to create client, %s",err.Error())
	}
	defer conn.Close()
	client := pbgen.NewMessageServiceClient(conn)
	in:=new (pbgen.RequestMessagesIn)
	in.ClientID="1"
	message, err := client.GetMessage(context.Background(), in)
	if err != nil{
		t.Fatalf("Error fectching messages %s",err.Error())
	}
	fmt.Println(message.MessageData)

}
