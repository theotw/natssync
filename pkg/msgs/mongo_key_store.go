/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TrustedLocation struct {
	LocationID string `json:"locationID" bson:"locationID"`
	PublicKey  []byte `json:"publicKey" bson:"publicKey"`
	MetaData   map[string]string `json:"metaData" bson:"metaData"`
}

type ServiceKeyData struct {
	ID         int    `json:"id" bson:"id"`
	LocationID string `json:"locationID" bson:"locationID"`
	PublicKey  []byte `json:"publicKey" bson:"publicKey"`
	PrivateKey []byte `json:"privateKey" bson:"privateKey"`
}

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
			Keys: bson.D{{"locationID", 1}},
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

func (m *MongoKeyStore) WriteKeyPair(locationID string, publicKey []byte, privateKey []byte) error {
	log.Tracef("Saving key pair for %s", locationID)
	keys := ServiceKeyData{
		ID: 1,
		LocationID: locationID,
		PublicKey: publicKey,
		PrivateKey: privateKey,
	}
	collection := m.getKeyPairCollection()
	_, err := collection.InsertOne(context.TODO(), keys)
	return err
}

func (m *MongoKeyStore) getServiceKeyPair() (*ServiceKeyData, error) {
	log.Trace("Reading key pair")
	var keypair ServiceKeyData
	collection := m.getKeyPairCollection()
	cur := collection.FindOne(context.TODO(), bson.D{{"id", 1}})
	if err := cur.Decode(&keypair); err != nil {
		return nil, fmt.Errorf("unable to get locationID from mongo: %s", err)
	}
	return &keypair, nil
}

func (m *MongoKeyStore) ReadKeyPair() ([]byte, []byte, error) {
	keypair, err := m.getServiceKeyPair()
	if err != nil {
		return nil, nil, err
	}
	return keypair.PublicKey, keypair.PrivateKey, nil
}

func (m *MongoKeyStore) RemoveKeyPair() error {
	log.Trace("Removing key pair")
	collection := m.getLocationsCollection()
	filter := bson.D{{"id", 1}}
	_, err := collection.DeleteOne(context.TODO(), filter)
	return err
}

func (m *MongoKeyStore) LoadLocationID() string {
	log.Trace("Getting location ID")
	keypair, err := m.getServiceKeyPair()
	if err != nil {
		log.Debugf("Unable to get locationID from mongo: %s", err)
		return ""
	}
	return keypair.LocationID
}

func (m *MongoKeyStore) WriteLocation(locationID string, buf []byte, metadata map[string]string) error {
	log.Tracef("Mongo set public key for '%s'", locationID)
	key := TrustedLocation{LocationID: locationID, PublicKey: buf, MetaData: metadata}
	collection := m.getLocationsCollection()
	_, err := collection.InsertOne(context.TODO(), key)
	return err
}

func (m *MongoKeyStore) ReadLocation(locationID string) ([]byte, map[string]string, error) {
	log.Tracef("Mongo get public key for '%s'", locationID)
	collection := m.getLocationsCollection()
	var location TrustedLocation
	cur := collection.FindOne(context.TODO(), bson.D{{"locationID", locationID}})
	err := cur.Decode(&location)
	return location.PublicKey, location.MetaData, err
}

func (m *MongoKeyStore) removeLocationData(locationID string, allowCloudMaster bool) error {
	if locationID == CLOUD_ID && !allowCloudMaster {
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
	return m.removeLocationData(CLOUD_ID, true)
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
		var location TrustedLocation
		err = cur.Decode(&location)
		if err != nil {
			return nil, err
		}
		locationIDs = append(locationIDs, location.LocationID)
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
