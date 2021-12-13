package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/httpsproxy/metrics"
	"github.com/theotw/natssync/pkg/httpsproxy/models"
	"github.com/theotw/natssync/pkg/httpsproxy/nats"
	"github.com/theotw/natssync/pkg/testing"
)

const (
	defaultLocationID = "proxy"
	locationIDEnvVar  = "DEFAULT_LOCATION_ID"
	proxyPortEnvVar   = "PROXY_PORT"
	defaultProxyPort  = "8080"
)

func getLocationIDFromEnv() string {
	return pkg.GetEnvWithDefaults(locationIDEnvVar, defaultLocationID)
}

func getProxyHostPort() string {
	port := pkg.GetEnvWithDefaults(proxyPortEnvVar, defaultProxyPort)
	return fmt.Sprintf(":%s", port)
}

type server struct {
	locationID   string
	natsClient   nats.ClientInterface
	unitTestMode bool
}

func NewServer() (*server, error) {
	locationID := getLocationIDFromEnv()
	natsClient, err := getInitializedNatsClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get nats client: %v", err)
	}

	server := NewServerDetailed(locationID, natsClient, false)
	return server, nil
}

func NewServerDetailed(locationID string, natsClient nats.ClientInterface, unitTestMode bool) *server {
	return &server{
		locationID:   locationID,
		natsClient:   natsClient,
		unitTestMode: unitTestMode,
	}
}

func getInitializedNatsClient() (nats.ClientInterface, error) {
	if err := models.InitNats(); err != nil {
		return nil, err
	}
	return models.GetNatsClient(), nil
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
func (s *server) RunHttpProxyServer(test bool) {

	s.configureNatsSyncLocationID()

	hostPort := getProxyHostPort()
	r := newRouter(s)
	srv := &http.Server{
		Addr:    hostPort,
		Handler: r,
	}

	log.WithField("hostPort", hostPort).Info("Starting server")

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

	metrics.InitProxyServerMetrics()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	log.Info("Server Started blocking on channel")
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	if test {
		testing.NotifyOnAppExitMessageGeneric(s.natsClient, quit)
	}
	<-quit

	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server Shutdown")
	}

	log.Info("Server exiting")
}
