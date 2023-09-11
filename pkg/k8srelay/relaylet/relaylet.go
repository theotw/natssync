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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	path2 "path"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	models "github.com/theotw/natssync/pkg/k8srelay/model"
	msgs "github.com/theotw/natssync/pkg/msgs"
	"github.com/theotw/natssync/pkg/natsmodel"
	"gopkg.in/yaml.v3"
)

type Relaylet struct {
	client *http.Client
	// caCert we always need the ca cert
	caCert string
	// clientKey is only needed if there is a client cert.  Most common, external test setup loaded from a kubeconfig
	clientKey  string
	clientCert string

	//clientToken is a token based auth, usually from an in pod SA account
	clientToken string

	serverURL string
}

func NewRelaylet() (*Relaylet, error) {
	x := new(Relaylet)
	err := x.init()
	return x, err
}
func (t *Relaylet) initFromInPodConfig() error {
	log.Infof("Loading in pod config")
	const (
		tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	)
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return errors.New("ErrNotInCluster")
	}
	t.serverURL = fmt.Sprintf("https://%s:%s", host, port)

	token, err := os.ReadFile(tokenFile)
	if err != nil {
		return err
	}
	t.clientToken = string(token)

	ca, err := os.ReadFile(rootCAFile)
	if err != nil {
		return err
	}
	t.caCert = base64.StdEncoding.EncodeToString(ca)
	return nil
}
func (t *Relaylet) initFromKubeConfig() error {
	log.Infof("Loading kubeconfig")
	path := os.Getenv("KUBECONFIG")
	if len(path) == 0 {
		home := os.Getenv("HOME")
		path = path2.Join(home, ".kube", "config")
	}
	log.Infof("Loading kubeconfig from %s", path)
	bits, err := os.ReadFile(path)
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
	t.serverURL = config.Clusters[0].Cluster.Server
	return nil
}
func (t *Relaylet) initK8SConfig() error {
	const namespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	k8sPod := true
	if _, err := os.Stat(namespaceFile); os.IsNotExist(err) {
		k8sPod = false
	}
	if k8sPod {
		return t.initFromInPodConfig()
	} else {
		return t.initFromKubeConfig()
	}
}

func (t *Relaylet) init() error {
	kcerr := t.initK8SConfig()
	if kcerr != nil {
		return kcerr
	}

	natsURL := os.Getenv("NATS_SERVER_URL")
	err := natsmodel.InitNats(natsURL, "relayserver", time.Minute*2)
	if err != nil {
		return err
	}
	caCert, err := base64.StdEncoding.DecodeString(t.caCert)
	if err != nil {
		log.WithError(err).Fatalf("Bad b4 %s", err.Error())
	}

	var clientCerts []tls.Certificate

	if len(t.clientCert) > 0 {
		clientCert, err := base64.StdEncoding.DecodeString(t.clientCert)
		if err != nil {
			log.WithError(err).Fatalf("Bad b4 %s", err.Error())
		}
		clientKey, err := base64.StdEncoding.DecodeString(t.clientKey)
		if err != nil {
			log.WithError(err).Errorf("Bad b4 %s", err.Error())
			return err
		}
		cert, cerr := tls.X509KeyPair(clientCert, clientKey)
		if cerr != nil {
			return cerr
		}
		clientCerts = make([]tls.Certificate, 1)
		clientCerts[0] = cert
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		return errors.New("unable to append pem ca cert")
	}

	t.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: clientCerts,
			},
		},
	}

	sub := msgs.MakeMessageSubject("*", models.K8SRelayRequestMessageSubjectSuffix)
	nc := natsmodel.GetNatsConnection()
	log.Infof("queue subscribing to %s", sub)
	_, err = nc.QueueSubscribe(sub, "k8srelay", func(msg *nats.Msg) {
		go t.DoCall(msg)
	})
	if err != nil {
		return errors.New("unable to QueueSubscribe for sub " + sub)
	}
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

	fullURL := fmt.Sprintf("%s%s", t.serverURL, req.Path)
	if len(req.QueryString) != 0 {
		fullURL = fmt.Sprintf("%s?%s", fullURL, req.QueryString)
	}

	inReader := bytes.NewReader(req.InBody)
	relayreq, err := http.NewRequest(req.Method, fullURL, inReader)
	if err != nil {
		panic(err)
	}

	for k, v := range req.Headers {
		relayreq.Header.Set(k, v)
	}
	if len(t.clientToken) > 0 {
		token := fmt.Sprintf("Bearer %s", t.clientToken)
		relayreq.Header.Set("Authorization", token)
	}

	if !req.Stream {
		t.callAPI(nc, nm, relayreq, respMsg, "", false)
	} else {
		go t.streamAPIMsgs(nc, nm, relayreq, respMsg, req.UUID)
	}
}

func (t *Relaylet) streamAPIMsgs(nc *nats.Conn, nm *nats.Msg, relayreq *http.Request, respMsg *models.CallResponse, requestUUID string) {
	log.Infof("starting streaming of API")
	t.callAPI(nc, nm, relayreq, respMsg, requestUUID, true)
}

func (t *Relaylet) callAPI(nc *nats.Conn, nm *nats.Msg, relayreq *http.Request, respMsg *models.CallResponse, requestUUID string, stream bool) {
	resp, err := t.client.Do(relayreq)
	if err != nil {
		respMsg.StatusCode = 502
		respMsg.AddHeader("Content-Type", "text/plain")
		errorstr := fmt.Sprintf("error from relay %s", err.Error())
		respMsg.OutBody = []byte(errorstr)
		respBits, err := json.Marshal(respMsg)
		if err != nil {
			log.WithError(err).Errorf("Unable to marshal response message %s", err.Error())
		}
		merr := nc.Publish(nm.Reply, respBits)
		if merr != nil {
			log.WithError(merr).Errorf("Error sending return message %s %s", nm.Reply, merr.Error())
		}
	} else {
		respMsg.StatusCode = resp.StatusCode
		log.WithField("URL", relayreq.URL.String()).
			Infof("Got resp status %d - len %d", resp.StatusCode, resp.ContentLength)
		for k, v := range resp.Header {
			respMsg.AddHeader(k, v[0])
		}

		streamStopChannel := make(chan int)
		if stream {
			sbMsgSub := msgs.MakeMessageSubject("*", models.K8SRelayRequestMessageSubjectSuffix+"."+requestUUID+".stopStreaming")
			sync, err := nc.SubscribeSync(sbMsgSub)
			if err != nil {
				log.WithError(err).Errorf("Unable to subscribe to %s response message %s", sbMsgSub, err.Error())
				return
			}
			go func() {
				for {
					_, err = sync.NextMsg(time.Minute * 5)
					if err != nil {
						if strings.Contains(err.Error(), "nats: timeout") {
							log.Warnf("timeout reading NextMsg %s, ignoring", err.Error())
						}
					} else {
						log.Info("got a request to stop streaming of API request")
						streamStopChannel <- 0
						return
					}
				}
			}()
		}

		for {
			if stream {
				select {
				case <-streamStopChannel:
					log.Info("stopping streaming of API request")
					return
				default:
				}
			}
			buf := make([]byte, 1024*1024)
			n, err := resp.Body.Read(buf)
			if err != nil {
				respMsg.LastMessage = true
				respMsg.OutBody = nil
				if err != io.EOF {
					log.WithError(err).Errorf("Error reading response stream %s", err.Error())
				}
			}
			if n < 0 {
				respMsg.OutBody = nil
			} else {
				respMsg.OutBody = buf[0:n]
			}

			respBits, err := json.Marshal(respMsg)
			if err != nil {
				log.WithError(err).Errorf("Unable to marshal response message %s", err.Error())
			}
			merr := nc.Publish(nm.Reply, respBits)
			if merr != nil {
				log.WithError(merr).Errorf("Error sending return message %s %s", nm.Reply, merr.Error())
			}
			log.Debugf("Receiving data size %d last message flag %v", n, respMsg.LastMessage)
			if respMsg.LastMessage {
				break
			}
		}
	}
}
