package msgs

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EncryptionKey struct {
	LocationID string `json:"locationID" bson:"locationID"`
	Bytes      []byte `json:"bytes" bson:"bytes"`
}

type location struct {
	ID         int    `json:"id" bson:"id"`
	LocationID string `json:"locationID" bson:"locationID"`
}

type MongoKeyStore struct {
	conn                  *mongo.Client
	options               *options.ClientOptions
	databaseName          string
	pubKeyCollectionName  string
	privKeyCollectionName string
	locIdCollectionName   string
}

func (m *MongoKeyStore) Init() error {
	var err error
	m.conn, err = mongo.Connect(context.TODO(), m.options)
	return err
}

func (m *MongoKeyStore) getPubKeyCollection() *mongo.Collection {
	return m.conn.Database(m.databaseName).Collection(m.pubKeyCollectionName)
}

func (m *MongoKeyStore) getPrivKeyCollection() *mongo.Collection {
	return m.conn.Database(m.databaseName).Collection(m.privKeyCollectionName)
}

func (m *MongoKeyStore) getLocIdCollection() *mongo.Collection {
	return m.conn.Database(m.databaseName).Collection(m.locIdCollectionName)
}

func (m *MongoKeyStore) SaveLocationID(locationID string) error {
	collection := m.getLocIdCollection()
	newLocation := location{1, locationID}
	_, err := collection.InsertOne(context.TODO(), newLocation)
	return err
}

func (m *MongoKeyStore) LoadLocationID() string {
	var locID location
	collection := m.getLocIdCollection()
	cur := collection.FindOne(context.TODO(), bson.D{{"id", 1}})
	if err := cur.Decode(&locID); err != nil {
		log.Debugf("Unable to get locationID from mongo: %s", err)
		return ""
	}
	return locID.LocationID
}

func (m *MongoKeyStore) ReadPrivateKeyData(locationID string) ([]byte, error) {
	var key EncryptionKey
	collection := m.getPrivKeyCollection()
	cur := collection.FindOne(context.TODO(), bson.D{{"locationID", locationID}})
	err := cur.Decode(&key)
	return key.Bytes, err
}

func (m *MongoKeyStore) WritePrivateKey(locationID string, buf []byte) error {
	log.Tracef("Mongo set private key for '%s'", locationID)
	key := EncryptionKey{LocationID: locationID, Bytes: buf}
	collection := m.getPrivKeyCollection()
	_, err := collection.InsertOne(context.TODO(), key)
	return err
}

func (m *MongoKeyStore) ReadPublicKeyData(locationID string) ([]byte, error) {
	log.Tracef("Mongo get public key for '%s'", locationID)
	collection := m.getPubKeyCollection()
	var instance EncryptionKey
	cur := collection.FindOne(context.TODO(), bson.D{{"locationID", locationID}})
	err := cur.Decode(&instance)
	return instance.Bytes, err
}

func (m *MongoKeyStore) WritePublicKey(locationID string, buf []byte) error {
	log.Tracef("Mongo set public key for '%s'", locationID)
	key := EncryptionKey{LocationID: locationID, Bytes: buf}
	collection := m.getPubKeyCollection()
	_, err := collection.InsertOne(context.TODO(), key)
	return err
}

func (m *MongoKeyStore) ListKnownClients() ([]string, error) {
	log.Trace("getting all known clients")
	collection := m.getPubKeyCollection()
	var instanceIDs []string

	cur, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var instance EncryptionKey
		err = cur.Decode(&instance)
		if err != nil {
			return nil, err
		}
		instanceIDs = append(instanceIDs, instance.LocationID)
	}

	if err = cur.Err(); err != nil {
		log.Fatal(err)
	}

	if err = cur.Close(context.TODO()); err != nil {
		log.Errorf("Error closing cursor: %s", err)
	}

	return instanceIDs, nil
}

func NewMongoKeyStore(mongoUri string) (*MongoKeyStore, error) {
	mongoUrl := fmt.Sprintf("mongodb://%s", mongoUri)
	log.Debugf("Connecting to mongo at %s", mongoUrl)

	keyStore := MongoKeyStore{
		options:               options.Client().ApplyURI(mongoUrl),
		databaseName:          "natssync",
		pubKeyCollectionName:  "publicKeys",
		privKeyCollectionName: "privateKeys",
		locIdCollectionName:   "locationIDs",
	}

	err := keyStore.Init()
	return &keyStore, err
}
