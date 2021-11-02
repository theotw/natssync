/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package mongo

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/types"
	"github.com/theotw/natssync/pkg/utils"
)

type MongoKeyStore struct {
	conn                    *mongo.Client
	options                 *options.ClientOptions
	databaseName            string
	keyPairCollectionName   string
	locationsCollectionName string
}

func (m *MongoKeyStore) initCollections() error {
	locCol := m.getLocationsCollection()
	_, err := locCol.Indexes().CreateOne(
		context.TODO(),
		mongo.IndexModel{
			Keys:    bson.D{{"id", 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return err
	}

	kpCol := m.getKeyPairCollection()
	_, err = kpCol.Indexes().CreateOne(
		context.TODO(),
		mongo.IndexModel{
			Keys:    bson.D{{"locationID", 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *MongoKeyStore) Init() error {
	var err error
	m.conn, err = mongo.Connect(context.TODO(), m.options)
	if err == nil {
		err = m.initCollections()
	}
	return err
}

func (m *MongoKeyStore) getKeyPairCollection() *mongo.Collection {
	return m.conn.Database(m.databaseName).Collection(m.keyPairCollectionName)
}

func (m *MongoKeyStore) getLocationsCollection() *mongo.Collection {
	return m.conn.Database(m.databaseName).Collection(m.locationsCollectionName)
}

func (m *MongoKeyStore) WriteKeyPair(locationData *types.LocationData) error {

	log.WithFields(log.Fields{
		"locationID": locationData.LocationID,
		"keyID": locationData.KeyID,
	}).Tracef("Saving key pair")

	collection := m.getKeyPairCollection()
	_, err := collection.InsertOne(context.TODO(), locationData)
	return err
}

func (m *MongoKeyStore) ReadKeyPair(keyID string) (*types.LocationData, error) {
	log.Trace("Reading key pair")
	if keyID == "" {
		var err error
		if keyID, err = m.GetLatestKeyID(); err != nil {
			return nil, err
		}
	}

	var keypair types.LocationData
	collection := m.getKeyPairCollection()
	cur := collection.FindOne(context.TODO(), bson.M{"keyID": keyID})
	if err := cur.Decode(&keypair); err != nil {
		return nil, fmt.Errorf("unable to get location for keyID %s from mongo: %s", keyID, err)
	}
	return &keypair, nil
}

func (m *MongoKeyStore) RemoveKeyPair(keyID string) error {
	log.Trace("Removing key pair")
	if keyID == "" {
		var err error
		if keyID, err = m.GetLatestKeyID(); err != nil {
			return err
		}
	}

	collection := m.getKeyPairCollection()
	filter := bson.M{"keyID": keyID}
	_, err := collection.DeleteOne(context.TODO(), filter)
	return err
}

func (m *MongoKeyStore) LoadLocationID(keyID string) string {
	log.Trace("Getting location ID")
	if keyID == "" {
		var err error
		if keyID, err = m.GetLatestKeyID(); err != nil {
			return ""
		}
	}

	keypair, err := m.ReadKeyPair(keyID)
	if err != nil {
		log.WithError(err).Debugf("Unable to get locationID from mongo")
		return ""
	}
	return keypair.GetLocationID()
}

func (m *MongoKeyStore) GetExistingKeys() ([]*utils.UUIDv1, error) {
	collection := m.getKeyPairCollection()
	filter := bson.M{}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := cur.Close(context.TODO()); err != nil {
			log.WithError(err).Error("failed to close mongo cursor")
		}
	}()

	existingKeys := make([]*utils.UUIDv1, 0)

	for cur.Next(context.TODO()) {
		locationData := &types.LocationData{}
		if err := cur.Decode(locationData); err != nil {
			return nil, err
		}
		idString := locationData.KeyID
		id, err := utils.ParseUUIDv1(idString)
		if err != nil {
			return nil, err
		}

		existingKeys = append(existingKeys, id)
	}
	return existingKeys, nil
}

func (m *MongoKeyStore) GetLatestKeyID() (string, error) {
	existingKeys, err := m.GetExistingKeys()
	if err != nil {
		return "", err
	}

	if len(existingKeys) == 0 {
		return "", fmt.Errorf("existing keys not found")
	}

	maxKey := existingKeys[0]
	for _, key := range existingKeys {
		if maxKey.GetCreationTime().Before(key.GetCreationTime()) {
			maxKey = key
		}
	}

	return maxKey.String(), nil
}

func (m *MongoKeyStore) WriteLocation(data types.LocationData) error {
	log.Tracef("Mongo set public key for '%s'", data.GetLocationID())

	collection := m.getLocationsCollection()
	_, err := collection.InsertOne(context.TODO(), data)
	if err != nil {
		log.WithError(err).Errorf("Error writing mongo record %v", data)
	}
	return err
}

func (m *MongoKeyStore) ReadLocation(locationID string) (*types.LocationData, error) {
	log.Tracef("Mongo get public key for '%s'", locationID)
	collection := m.getLocationsCollection()
	var location types.LocationData
	cur := collection.FindOne(context.TODO(), bson.D{{"locationID", locationID}})
	err := cur.Decode(&location)
	return &location, err
}

func (m *MongoKeyStore) removeLocationData(locationID string, allowCloudMaster bool) error {
	if locationID == pkg.CLOUD_ID && !allowCloudMaster {
		log.Errorf("Removing default cloud location ID")
		err := errors.New("unable to remove cloud master location")
		return err
	}

	log.Tracef("Mongo remove location for '%s'", locationID)
	_, err := m.getLocationsCollection().DeleteOne(context.TODO(), bson.D{{"locationID", locationID}})
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoKeyStore) RemoveLocation(locationID string) error {
	return m.removeLocationData(locationID, false)
}

func (m *MongoKeyStore) RemoveCloudMasterData() error {
	return m.removeLocationData(pkg.CLOUD_ID, true)
}

func (m *MongoKeyStore) ListKnownClients() ([]string, error) {
	log.Trace("getting all known clients")
	collection := m.getLocationsCollection()
	var locationIDs []string

	cur, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var location types.LocationData
		err = cur.Decode(&location)
		if err != nil {
			return nil, err
		}
		locationIDs = append(locationIDs, location.GetLocationID())
	}

	if err = cur.Err(); err != nil {
		log.Fatal(err)
	}

	if err = cur.Close(context.TODO()); err != nil {
		log.Errorf("Error closing cursor: %s", err)
	}

	return locationIDs, nil
}

func NewMongoKeyStore(mongoUri string) (*MongoKeyStore, error) {
	mongoUrl := fmt.Sprintf("mongodb://%s", mongoUri)
	log.Debugf("Connecting to mongo at %s", mongoUrl)

	keyStore := MongoKeyStore{
		options:                 options.Client().ApplyURI(mongoUrl),
		databaseName:            "natssync",
		keyPairCollectionName:   "keypair",
		locationsCollectionName: "locations",
	}

	err := keyStore.Init()
	return &keyStore, err
}
