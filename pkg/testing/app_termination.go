package testing

import (
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	nats2 "github.com/theotw/natssync/pkg/httpsproxy/nats"
	"os"
)

type NatsConnectionInterface interface {
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
}

const AppExitTopic = "natssync.testing.exitapp"

// NotifyOnAppExitMessage subscribes to the AppExitTopic for the given NATS client. When a message is received, an os.Interrupt
// signal is sent using the given channel. This can then be used to exit the app 'gracefully', which is required for collecting code
// coverage reports generated via 'go test'.
//
// WARNING: this function should only be used during testing (generally for the purposes of collecting coverage, as explained above).
func NotifyOnAppExitMessage(natsConnection NatsConnectionInterface, quitChannel chan os.Signal) {
	log.Warn("A testing-only function is being called. If you see this in production, something is very wrong!")

	_, err := natsConnection.Subscribe(AppExitTopic, func(msg *nats.Msg) {
		log.Info("Termination command received via NATS, sending interrupt signal...")
		quitChannel <- os.Interrupt
	})

	if err != nil {
		log.WithError(err).Fatal("failed to subscribe to the app exit topic")
	}

	log.Infof("Succesfully subscribed to the app exit topic. To exit the app gracefully, send a NATS message to: %s", AppExitTopic)
}

// NotifyOnAppExitMessageGeneric does the same thing as NotifyOnAppExitMessage but allows using the generic interface defined in this
// repository.
func NotifyOnAppExitMessageGeneric(client nats2.ClientInterface, quitChannel chan os.Signal) {
	log.Warn("A testing-only function is being called. If you see this in production, something is very wrong!")

	_, err := client.Subscribe(AppExitTopic, func(msg *nats2.Msg) {
		log.Info("Termination command received via NATS, sending interrupt signal...")
		quitChannel <- os.Interrupt
	})

	if err != nil {
		log.WithError(err).Fatal("failed to subscribe to the app exit topic")
	}

	log.Infof("Succesfully subscribed to the app exit topic. To exit the app gracefully, send a NATS message to: %s", AppExitTopic)
}
