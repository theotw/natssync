/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package apps

import (
	"crypto/tls"
	"github.com/theotw/natssync/pkg/cloudserver"
	"net/http"
	"testing"
	"time"
)

func TestBridgeServer(t *testing.T) {
	cloudserver.RunBridgeServerApp(true)
}
func makeHttpCall(url string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Timeout:   5 * time.Second,
		Transport: tr,
	}
	client.Head(url)
}
