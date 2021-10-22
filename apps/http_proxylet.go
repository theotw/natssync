/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/httpsproxy"
	models2 "github.com/theotw/natssync/pkg/httpsproxy/models"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
	"io/ioutil"
	"net/http"
	"runtime"
)

func runHttpAPI(m *nats.Msg, nc *nats.Conn, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'", i, m.Subject, string(m.Data))
	var req server.HttpApiReqMessage
	jsonerr := json.Unmarshal(m.Data, &req)
	var resp server.HttpApiResponseMessage
	if jsonerr != nil {
		log.Errorf("Error decoding http message %s", jsonerr.Error())
		resp.HttpStatusCode = 502
		resp.RespBody = jsonerr.Error()
		nc.Publish(m.Reply, []byte("ack"))
		nc.Flush()
		return
	}
	urlToUse := fmt.Sprintf("http://%s%s", req.Target, req.HttpPath)
	reader := bytes.NewReader(req.Body)
	localreq, _ := http.NewRequest(req.HttpMethod, urlToUse, reader)
	for _, x := range req.Headers {
		localreq.Header.Add(x.Key, x.Values[0])
	}

	localresp, httperr := http.DefaultClient.Do(localreq)
	if httperr != nil {
		log.Errorf("Error decoding http message %s", httperr.Error())
		resp.HttpStatusCode = 502
		resp.RespBody = httperr.Error()
	} else {
		resp.HttpStatusCode = localresp.StatusCode
		bodybits, bodyerr := ioutil.ReadAll(localresp.Body)
		if bodyerr == nil {
			resp.RespBody = string(bodybits)
		} else {
			log.Errorf("Got an error reading a resp body from http api %s %s", req.Target, bodyerr.Error())
			resp.RespBody = bodyerr.Error()
		}
		resp.Headers = make(map[string]string)
		for k, v := range localresp.Header {
			if len(v) > 0 {
				resp.Headers[k] = v[0]
			} else {
				resp.Headers[k] = ""
			}
		}
	}
	respBits, jsonerr := json.Marshal(&resp)
	if jsonerr != nil {
		log.Errorf("Error encoding resp body bits %s", jsonerr.Error())
	}
	nc.Publish(m.Reply, respBits)
	nc.Flush()

}
func ConfigureDefaultTransport() {
	// allows calls to https
	dt := http.DefaultTransport
	switch dt.(type) {
	case *http.Transport:
		if dt.(*http.Transport).TLSClientConfig == nil {
			dt.(*http.Transport).TLSClientConfig = &tls.Config{}
		}
		dt.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
	}
}
func main() {
	log.Info("Version %s", pkg.VERSION)
	logLevel := httpproxy.GetEnvWithDefaults("LOG_LEVEL", "debug")
	level, levelerr := log.ParseLevel(logLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)
	err := models2.InitNats()
	if err != nil {
		log.Fatalf("Unable to init NATS %s", err.Error())
	}
	nc := models2.GetNatsClient()
	_, err = nc.Subscribe(server.ResponseForLocationID, func(msg *nats.Msg) {
		locationID := string(msg.Data)
		httpproxy.SetMyLocationID(locationID)
		log.Infof("Using location ID %s", locationID)

	})
	if err != nil {
		log.Fatalf("Unable to talk to NATS %s", err.Error())
	}
	nc.Publish(server.RequestForLocationID, []byte(""))

	ConfigureDefaultTransport()
	err = models2.InitNats()
	if err != nil {
		log.Fatalf("Unable to connect to NATS %s", err.Error())
	}
	locationID := httpproxy.GetMyLocationID()
	subj := httpproxy.MakeMessageSubject(locationID, httpproxy.HTTP_PROXY_API_ID)
	i := 0
	nc.Subscribe(subj, func(msg *nats.Msg) {
		if string(msg.Data) != "" {
			i += 1
			runHttpAPI(msg, nc, i)
		}
	})
	log.Printf("Listening on [%s]", subj)

	conReqSubject := httpproxy.MakeMessageSubject(locationID, httpproxy.HTTPS_PROXY_CONNECTION_REQUEST)
	nc.Subscribe(conReqSubject, func(msg *nats.Msg) {
		models2.HandleConnectionRequest(msg, locationID)
	})

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on [%s]", conReqSubject)
	nc.Flush()

	//if *showTime {
	//	log.SetFlags(log.LstdFlags)
	//}

	runtime.Goexit()
}
