package testing

import (
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"os"
)

type NatsConnectionInterface interface {
	Subscribe(topic string, msgHandler nats.MsgHandler) (*nats.Subscription, error)
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
