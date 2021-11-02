package cloudserver

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/persistence"
)

const (
	defaultCertRotationTimeout = 24 * time.Hour
	certRotationTimeoutEnvKey  = "CERT_ROTATION_TIMEOUT"
)

type certMiddleware struct {
	timeout     time.Duration
	persistence persistence.LocationKeyStore
}

func NewCertMiddleware(persistence persistence.LocationKeyStore) *certMiddleware {
	certRotationTimeout := defaultCertRotationTimeout

	if timeoutString, exists := os.LookupEnv(certRotationTimeoutEnvKey); exists {
		if timeout, err := time.ParseDuration(timeoutString); err != nil {
			log.WithError(err).Errorf("failed to parse timeout from environment")
		} else {
			certRotationTimeout = timeout
		}
	}
	log.Infof("setting cert rotation timeout to %v", certRotationTimeout.String())
	return NewCertMiddlewareDetailed(certRotationTimeout, persistence)
}

func NewCertMiddlewareDetailed(timeout time.Duration, persistence persistence.LocationKeyStore) *certMiddleware {
	return &certMiddleware{
		timeout:     timeout,
		persistence: persistence,
	}
}

func (c *certMiddleware) Enforce(ginContext *gin.Context) {

	clientID := ginContext.Param("premid")

	data, err := c.persistence.ReadLocation(clientID)
	if err != nil {
		log.Warning("failed to read location data from persistence")
	} else {

		timePeriodSinceLastCertRotation := time.Now().Sub(data.GetLastKeyPairRotation())

		if timePeriodSinceLastCertRotation >= c.timeout || data.GetForceKeypairRotation() {
			log.WithField("clientID", clientID).Infof("sending out cert rotation request")
			ginContext.AbortWithStatusJSON(pkg.StatusCertificateError, "")
			return
		}
	}

	ginContext.Next()
}
