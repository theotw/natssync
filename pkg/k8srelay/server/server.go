package server

import (
	"context"
	"encoding/base64"
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
const myserverert = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVSRENDQXl5Z0F3SUJBZ0lKQUl6RlRkdmQyYndtTUEwR0NTcUdTSWIzRFFFQkN3VUFNSUdETVFzd0NRWUQKVlFRR0V3SkZRekVQTUEwR0ExVUVDQXdHVFdGdVlXSnBNUTR3REFZRFZRUUhEQVZOWVc1MFlURVVNQklHQTFVRQpDZ3dMUlc1bmFXNWxaWEpwYm1jeEREQUtCZ05WQkFzTUEwUmxkakVOTUFzR0ExVUVBd3dFY205dmRERWdNQjRHCkNTcUdTSWIzRFFFSkFSWVJiV0Z6YjI1aVFHNWxkR0Z3Y0M1amIyMHdIaGNOTWpNd016SXlNVGN4TmpJeFdoY04KTWpVd05qSTBNVGN4TmpJeFdqQm1NUXN3Q1FZRFZRUUdFd0pGUXpFUE1BMEdBMVVFQ0F3R1RXRnVZV0pwTVE0dwpEQVlEVlFRSERBVk5ZVzUwWVRFVU1CSUdBMVVFQ2d3TFJXNW5hVzVsWlhKcGJtY3hEREFLQmdOVkJBc01BMFJsCmRqRVNNQkFHQTFVRUF3d0piRzlqWVd4b2IzTjBNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUIKQ2dLQ0FRRUFxWFBBcVMvNVRZNWdmbnBZREVNd3l1bmpTYndFaU5KS01KMDZSSnF3ZGdadGtjaDN0azh2d0N4dgpVa2JST0phZitzTWx6SG1TQTNwZXNOTUtncXVTZG5aKy83UzhDZTVkWXVoVDZ2anFzNjZIZGllK2cySW05UTRUCjhHTUZxaXdZd05TdytwYnk5UVVCRzIxc1hpTUFEN3dvQkhJS2ZGa01paVFZTHdBVjhLOFozSEhLNWdFdU5rTXoKNjViZTF2MTlaMkZJZzJCeGxEOGQvckxGbUIxNW10eVZQaGtKTzhWTjJxbnNBb3dXVC9MaVcvQk15RGlxS3Y5agpIV3ZqUG1QSDQ1T0p5Yi9ERW96WWF5SGhTOXRhbmc2REFHUytsMGNmdTVaTUlrRloyNHdwVFR6VHkrSk0xaGVECmVZNys0dUlzZytpbXg3eUxKcEpLMjZPQ0dHalJ3d0lEQVFBQm80SFdNSUhUTUlHaUJnTlZIU01FZ1pvd2daZWgKZ1lta2dZWXdnWU14Q3pBSkJnTlZCQVlUQWtWRE1ROHdEUVlEVlFRSURBWk5ZVzVoWW1reERqQU1CZ05WQkFjTQpCVTFoYm5SaE1SUXdFZ1lEVlFRS0RBdEZibWRwYm1WbGNtbHVaekVNTUFvR0ExVUVDd3dEUkdWMk1RMHdDd1lEClZRUUREQVJ5YjI5ME1TQXdIZ1lKS29aSWh2Y05BUWtCRmhGdFlYTnZibUpBYm1WMFlYQndMbU52YllJSkFJeUwKQlFMNGQwelRNQWtHQTFVZEV3UUNNQUF3Q3dZRFZSMFBCQVFEQWdUd01CUUdBMVVkRVFRTk1BdUNDV3h2WTJGcwphRzl6ZERBTkJna3Foa2lHOXcwQkFRc0ZBQU9DQVFFQVBtMmRiNUp5dm5FRVljWUFaeWNqQlQwYVZ3dGNmL2dQCnBYOWpFUmF5R0R1WHpFY3V1bWI0M0dwdjI1NlJ0S2d1azNyYXBtNlgraS80ZkR6RFdUZVgwWEdMbFBaZTRyTTcKUHNpKzJTSFZTaVJRbmxrWVZSM3VKZmZEc1JsMDVoNXhhcHhScEFsZk5kN2RSa092NURYbFRyZjNnTWh2T3d0Nwo1YVRJMVQwZEhrVjV1VFBpcU5nK3V1c2E1WlVjajFScTAvWSt6S0tQMXhGVjM0R0I1T1VmSUZGWGRwenRVby9QCjRuZGFyOFM2cndYWGhtR1hkMmM3cDFjT2kyZHZMbUtGbjNKUitucWVSck0zSTNtQ1NwSkVGQ1N4Qk5JTkxNcjIKYm83VVN5dkUzL0NiNlJzVEJtZ2hGTE5ORThUYWpnV2duVWZZMnVkMnh1dlk5UW1LL3RRVjFBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="

const myserverkey = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBcVhQQXFTLzVUWTVnZm5wWURFTXd5dW5qU2J3RWlOSktNSjA2Ukpxd2RnWnRrY2gzCnRrOHZ3Q3h2VWtiUk9KYWYrc01sekhtU0EzcGVzTk1LZ3F1U2RuWisvN1M4Q2U1ZFl1aFQ2dmpxczY2SGRpZSsKZzJJbTlRNFQ4R01GcWl3WXdOU3crcGJ5OVFVQkcyMXNYaU1BRDd3b0JISUtmRmtNaWlRWUx3QVY4SzhaM0hISwo1Z0V1TmtNejY1YmUxdjE5WjJGSWcyQnhsRDhkL3JMRm1CMTVtdHlWUGhrSk84Vk4ycW5zQW93V1QvTGlXL0JNCnlEaXFLdjlqSFd2alBtUEg0NU9KeWIvREVvellheUhoUzl0YW5nNkRBR1MrbDBjZnU1Wk1Ja0ZaMjR3cFRUelQKeStKTTFoZURlWTcrNHVJc2craW14N3lMSnBKSzI2T0NHR2pSd3dJREFRQUJBb0lCQUZsOWN5c20xZCs2cUlWRApNWXJRVlUxa2RnK3p4eVZIQWIxbzI2UHRtZkhLOVVTL2ZWRi93blVZUW5aT1JpSS9raCtKdmtXZGtwcFpudlo5CmpoaHlhZmc4SGxnRzZDUEtpZkU1UjFCWnd3Ry8wM1I0Q3VveUJPYjRWMWxsd2xFYjFyckgyT3VPbXFNQjBKTGUKbUJPaklsNHMvV2xUbk93TXowMkpRR2hhQUR4S01TbFRYbGxuQ0s3ZVErQXhpdWtINDBoS3FWN0o0Zm9oeWhuMQpTUE9jS2lWb1grVmNTOGVUWklmUXdGaHhHZlVsNkY1VEpaWFpRdjhIVkEwWlBvRE9KU1FJTVBHQjlyb0lGMWlaCndsSnJYWTQ4WTg0ckRlNEI3b3RCenFHRGxhRDNPbENBYWdRTUw0QWNmWExTalhoUWpZS0gvK2I3VCtsT2VPc0gKako1cmZRRUNnWUVBMWVDdXB6VitNSGhxckZkdlNiYnBoSjQ2djAxOEJ2aEhQUFZ2VjdlYWU0RjZLWS9oSkMvVQpwSERWMFp5c0xkbHZMM0p5QTFRdTZEbkJkc1dFdXAzbzJDM2VVTFdXbWdZblRhZ0xtYW83UUNMV3VyY2toZThFClJyYmxuU24yVTdIUldENUZxL2JxZktQNXU1dVlnYS80aytsR3haVml0NFVFd1ZoZHhxaFNsNEVDZ1lFQXl0TTMKMmQvTGJLVXplcENkc2tOajd0ZnhwV3JiVGtJbG9jeE1kQjdUaHdMcTNtNkN6RWJZTDUwaXl3L09SWmtGSkIwYQpLeTN4VS8rdGJMeGxuTEU0bVAzMFVwZHJiSDFiYzA5Z0ltVmw2SFJUSkVXdDdPVUFqSVBYTm04czMvcHVKSHhKCm9Vc3ZSb2V4bmlCV3hrRklvS2FhZnFrbTJkRStIS21MdUd1WXEwTUNnWUIyVVNCdGNlTklMeVZjQjlhUjRmVlgKSHkyQ3JQdkM0MUNOZ1gxQitsa2tuK0VUNHZ0NnlGY0xUVHlNQSs4Z1Fod0hGSG5NSzZMelp4Z0dlNGhNc0pTaQpHdVhVb2xBWkR2UnBPbUNJZHFybWRSOXpGV1BJRUF5K2plbUNRemQ0MzNMZkxUdmZ3TzNCVy9rSWR6QXI5a3crCmp4dE9yTEI2czhTSXJUamJjRHlZZ1FLQmdRQ09xODBad1ViREFlSVlVU25jZjNNSVMzWjd1WkxTbGMwSzV6N1EKWCs3RGhkWFk2VHV3bmhUc1NVaDBOb1lPaHZrSzBqM2FLZE1jRnpuU3h5TmkrWGFxaDlrWlQ5SU0ycEU5cDVRawpIZGQxa0gzN2dkZzZUMHYzaTdZVFlGamNwTGhkaWQveFNZWFR3eGU1OTEyU3o5RmEwRmJVeWYrK01HWFVtNTdnCjl5N0pLUUtCZ1FDek5DU1R5ODJhbEdGN2ErMDYwb1V6c29Sa3JqcjRXZWZFbHM0RnZGdDRUMmROMW9lS0QrcTMKQnhWaE82VDBnVTYxdUZ0NysxY1hYeHltUm94MWxucm96ckFLaUtWV3dDYTA0bHNCZnRNQTNzVzVUb3JuZHc5WQpGMjdESStZQzlNelJIYVpkQVZvMTl3VEVES3dCY004Uzl4ek1ZTUhETFBhUUE5cndBeEh3Nnc9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="

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

	tmp := os.Getenv("TMPDIR")
	keyFile := path.Join(tmp, "keyfile")
	f, err := os.Create(keyFile)
	if err != nil {
		panic(err)
	}
	bits, decoderror := base64.StdEncoding.DecodeString(myserverkey)
	if decoderror != nil {
		panic(decoderror)
	}
	f.Write(bits)
	f.Close()
	certFile := path.Join(tmp, "certfile")
	f, err = os.Create(certFile)
	if err != nil {
		panic(err)
	}
	bits, decoderror = base64.StdEncoding.DecodeString(myserverert)
	if decoderror != nil {
		panic(decoderror)
	}

	f.Write([]byte(bits))
	f.Close()

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
