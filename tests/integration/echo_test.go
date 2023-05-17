/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/msgs"
)

func GetNatsSyncClientLocationID(t *testing.T) string {
	url := pkg.GetEnvWithDefaults("syncclient_url", "http://localhost:30281")
	regURL := fmt.Sprintf("%s/bridge-client/1/register", url)
	resp, err := http.Get(regURL)
	if !assert.Nil(t, err) {
		t.Fatalf("Unable to get reg code %s ", err.Error())
	}
	if !assert.Equal(t, resp.StatusCode, 200) {
		t.Fatalf("Unable to get reg code %s ", resp.Status)
	}
	data := make(map[string]interface{})
	all, err := ioutil.ReadAll(resp.Body)
	if !assert.Nil(t, err) {
		t.Fatalf("Unable to read body %s ", err.Error())
	}

	err = json.Unmarshal(all, &data)
	if !assert.Nil(t, err) {
		t.Fatalf("Unable to unmarsell body %s ", err.Error())
	}
	locationOb := data["locationID"]
	return locationOb.(string)
}

func TestEcho(t *testing.T) {
	natsURL := pkg.GetEnvWithDefaults("natsserver_url", "nats://localhost:30221")

	nc, err := nats.Connect(natsURL)
	if !assert.Nil(t, err) {
		t.Fatalf("Unable to connect to NATS %s ", err.Error())
	}

	locationID := GetNatsSyncClientLocationID(t)
	subject := msgs.MakeEchoSubject(locationID)
	replySubject := msgs.MakeNBReplySubject()
	replyListenSub := fmt.Sprintf("%s.*", replySubject)
	sync, err := nc.SubscribeSync(replyListenSub)
	if err != nil {
		t.Fatalf("Error subscribing: %e", err)
	}

	err = nc.PublishRequest(subject, replySubject, []byte("ping"))
	if err != nil {
		t.Fatalf("Error on ping publish %s", err.Error())
	}
	for {
		msg, err := sync.NextMsg(1 * time.Minute)

		if err != nil {
			t.Fatalf("Error from sync.NextMsg %s", err.Error())
		} else {
			fmt.Printf("Message received [%s]: %s", msg.Subject, string(msg.Data))
			if strings.HasSuffix(msg.Subject, msgs.ECHOLET_SUFFIX) {
				break
			}
		}
	}

}
