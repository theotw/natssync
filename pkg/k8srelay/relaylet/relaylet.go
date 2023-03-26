/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package relaylet

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	models "github.com/theotw/natssync/pkg/k8srelay/model"
	msgs "github.com/theotw/natssync/pkg/msgs"
	"github.com/theotw/natssync/pkg/natsmodel"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	path2 "path"
	"time"
)

type Relaylet struct {
	client     *http.Client
	caCert     string
	clientKey  string
	clientCert string
}

func NewRelaylet() (*Relaylet, error) {
	x := new(Relaylet)
	err := x.init()
	return x, err
}
func (t *Relaylet) initFromKubeConfig() error {
	path := os.Getenv("KUBECONFIG")
	if len(path) == 0 {
		home := os.Getenv("HOME")
		path = path2.Join(home, ".kube", "config")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	bits, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	var config models.KubeConfigCluster
	err = yaml.Unmarshal(bits, &config)
	if err != nil {
		return err
	}
	t.clientKey = config.Users[0].User.ClientKeyData
	t.clientCert = config.Users[0].User.ClientCertificateData
	t.caCert = config.Clusters[0].Cluster.CertificateAuthorityData
	return nil
}

func (t *Relaylet) init() error {
	//todo, need to make this optional and initialize from pod env vars
	kcerr := t.initFromKubeConfig()
	if kcerr != nil {
		return kcerr
	}

	natsURL := os.Getenv("NATS_URL")
	err := natsmodel.InitNats(natsURL, "relayserver", time.Minute*2)
	if err != nil {
		return err
	}
	caCert, err := base64.StdEncoding.DecodeString(t.caCert)
	if err != nil {
		log.WithError(err).Fatalf("Bad b4 %s", err.Error())
	}

	clientCert, err := base64.StdEncoding.DecodeString(t.clientCert)
	if err != nil {
		log.WithError(err).Fatalf("Bad b4 %s", err.Error())
	}
	clientKey, err := base64.StdEncoding.DecodeString(t.clientKey)
	if err != nil {
		log.WithError(err).Fatalf("Bad b4 %s", err.Error())
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		panic("not ok")
	}
	cert, cerr := tls.X509KeyPair(clientCert, clientKey)
	if cerr != nil {
		panic("bad keypait")
	}

	t.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	sub := msgs.MakeMessageSubject("*", models.K8SRelayRequestMessageSubjectSuffix)
	nc := natsmodel.GetNatsConnection()
	nc.Subscribe(sub, func(msg *nats.Msg) {
		t.DoCall(msg)
	})
	return nil
}

const debug = false

func (t *Relaylet) DoCall(nm *nats.Msg) {
	respMsg := models.NewCallResponse()
	nc := natsmodel.GetNatsConnection()
	var req models.CallRequest
	err := json.Unmarshal(nm.Data, &req)
	if err != nil {
		respMsg.StatusCode = 502
		respMsg.AddHeader("Content-Type", "text/plain")
		errorstr := fmt.Sprintf("error from relay %s", err.Error())
		respMsg.OutBody = []byte(errorstr)
		respBits, err := json.Marshal(respMsg)
		if err != nil {
			log.WithError(err).Errorf("Unable to marshal response message %s", err.Error())
		}
		nc.Publish(nm.Reply, respBits)
		return
	}

	//TODO need to read this from Env Vars
	hostBase := "https://kubernetes.docker.internal:6443"
	fullURL := fmt.Sprintf("%s%s", hostBase, req.Path)

	inReader := bytes.NewReader(req.InBody)
	relayreq, err := http.NewRequest(req.Method, fullURL, inReader)
	if err != nil {
		panic(err)
	}

	resp, err := t.client.Do(relayreq)
	var respBody []byte

	if err != nil {
		respMsg.StatusCode = 502
		respMsg.AddHeader("Content-Type", "text/plain")
		errorstr := fmt.Sprintf("error from relay %s", err.Error())
		respMsg.OutBody = []byte(errorstr)
	} else {
		respMsg.StatusCode = resp.StatusCode
		log.Infof("Got resp status %d - len %d", resp.StatusCode, resp.ContentLength)
		for k, v := range resp.Header {
			respMsg.AddHeader(k, v[0])
		}
		respBody, err = io.ReadAll(resp.Body)
		if err == nil {
			respMsg.OutBody = respBody
		}
	}
	respBits, err := json.Marshal(respMsg)
	if err != nil {
		log.WithError(err).Errorf("Unable to marshal response message %s", err.Error())
	}
	nc.Publish(nm.Reply, respBits)

}
