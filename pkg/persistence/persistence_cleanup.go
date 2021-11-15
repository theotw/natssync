package persistence

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg/utils"
)

const (
	cleanupInterval = 1 * time.Hour
	cleanupTTL      = 5 * time.Minute
)

type CleanupKeysInterface interface {
	GetExistingKeys() ([]*utils.UUIDv1, error)
	GetLatestKeyID() (string, error)
	RemoveKeyPair(keyID string) error
}

type reaper struct {
	store           CleanupKeysInterface
	cleanupInterval time.Duration
	cleanupTTL      time.Duration
}

func newReaper(store CleanupKeysInterface) *reaper {
	return newReaperDetailed(store, cleanupInterval, cleanupTTL)
}

func newReaperDetailed(
	store CleanupKeysInterface,
	cleanupInterval time.Duration,
	cleanupTTL time.Duration,
) *reaper {

	return &reaper{
		store:           store,
		cleanupInterval: cleanupInterval,
		cleanupTTL:      cleanupTTL,
	}
}

func (r *reaper) RunCleanupJob(ctx context.Context) {

	ticker := time.NewTicker(r.cleanupInterval)

	go func() {
		log.Infof("setting up cleanup interval: %v", r.cleanupInterval)
		for {
			select {
			case <-ticker.C:
				log.Infof("running cleanup")
				r.cleanupOldKeys()
				log.Infof("cleanup complete")

			case <-ctx.Done():
				ticker.Stop()
				log.Infof("shutting down auto cleanup")
				return
			}
		}
	}()
}

func (r *reaper) cleanupOldKeys() {
	existingKeys, err := r.store.GetExistingKeys()
	if err != nil {
		log.WithError(err).Error("failed to get existing keys")
	}

	latestKeyID, err := r.store.GetLatestKeyID()
	if err != nil {
		log.WithError(err).Error("failed to get the latest key ID")
	}

	latestKey, err := utils.ParseUUIDv1(latestKeyID)
	if err != nil {
		log.WithError(err).Error("failed to parse the latest key ID")
	}

	for _, key := range existingKeys {

		// do not delete the latest key or any key created after the latest key
		if key.String() == latestKeyID || key.GetCreationTime().After(latestKey.GetCreationTime()) {
			continue
		}

		if time.Now().Sub(key.GetCreationTime()) >= r.cleanupTTL {
			if err = r.store.RemoveKeyPair(key.String()); err != nil {
				log.WithError(err).WithField("keyID", key).Error("failed to delete expired key pair")
			} else {
				log.WithField("keyID", key).Debugf("successfully deleted expired key")
			}
		}
	}
}
