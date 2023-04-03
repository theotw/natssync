package server

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/theotw/natssync/pkg/natsmodel"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
)

const (
	defaultLocationID = "relay"
	locationIDEnvVar  = "DEFAULT_LOCATION_ID"
	proxyPortEnvVar   = "RELAY_PORT"
	defaultRelayPort  = "8080"
)

var myserverert = "dev"

var myserverkey = "dev"

func getLocationIDFromEnv() string {
	return pkg.GetEnvWithDefaults(locationIDEnvVar, defaultLocationID)
}

func getRelayPort() string {
	port := pkg.GetEnvWithDefaults(proxyPortEnvVar, defaultRelayPort)
	return fmt.Sprintf(":%s", port)
}

type server struct {
	locationID   string
	natsClient   *nats.Conn
	unitTestMode bool
}

func NewServer() (*server, error) {
	locationID := getLocationIDFromEnv()
	natsURL := os.Getenv("NATS_SERVER_URL")
	err := natsmodel.InitNats(natsURL, "relayserver", time.Minute*2)
	if err != nil {
		return nil, err
	}

	server := newRelayServer(locationID, false)
	return server, nil
}

func newRelayServer(locationID string, unitTestMode bool) *server {
	return &server{
		locationID:   locationID,
		unitTestMode: unitTestMode,
	}
}

func (s *server) configureNatsSyncLocationID() {

	if _, err := s.natsClient.Subscribe(ResponseForLocationID, func(msg *nats.Msg) {
		s.locationID = string(msg.Data)
		log.WithField("locationID", s.locationID).Info("Using location ID")

	}); err != nil {
		log.WithError(err).Fatalf("Unable to talk to NATS")
	}

	if err := s.natsClient.Publish(RequestForLocationID, []byte("")); err != nil {
		log.WithError(err).Errorf("failed to send request for locationID")
	}
}

// Run - configures and starts the web server
func (s *server) RunRelayServer(test bool) error {
	natsurl := os.Getenv("NATS_SERVER_URL")
	err := natsmodel.InitNats(natsurl, "relay server", 2*time.Minute)
	if err != nil {
		return err
	}
	s.natsClient = natsmodel.GetNatsConnection()
	s.configureNatsSyncLocationID()

	tmp := pkg.GetEnvWithDefaults("CERT_DIR", "out/")
	keyFile := path.Join(tmp, "k8srelay.key")
	_, err = os.Stat(keyFile)
	if err != nil {
		log.WithError(err).Errorf("Unable to find server key file at path %s", keyFile)
		return err
	}
	certFile := path.Join(tmp, "k8srelay.crt")
	_, err = os.Stat(certFile)
	if err != nil {
		log.WithError(err).Errorf("Unable to find server cert file at path %s", certFile)
		return err
	}

	//hostPort := getRelayPort()
	hostPort := ":8443"
	r := newRouter(s)
	srv := &http.Server{
		Addr:    hostPort,
		Handler: r,
	}

	log.WithField("hostPort", hostPort).Info("Starting server")

	go func() {
		// service connections
		//err := srv.ListenAndServe()
		certFile := certFile
		keyFile := keyFile
		err := srv.ListenAndServeTLS(certFile, keyFile)
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

	//metrics.InitProxyServerMetrics()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	log.Info("Server Started blocking on channel")
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	if test {
		//TODO testing.NotifyOnAppExitMessageGeneric(s.natsClient, quit)
	}
	<-quit

	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server Shutdown")
	}

	log.Info("Server exiting")
	return nil
}