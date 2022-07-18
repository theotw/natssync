/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package grpc

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	natssyncgrpc "github.com/theotw/natssync/pkg/grpc"
	"github.com/theotw/natssync/pkg/pbgen"
	grpc "google.golang.org/grpc"
	"net"
	"testing"
)

func TestGprcServer(t *testing.T) {
	impl := new(natssyncgrpc.MessageServerImpl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:8082"))
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
