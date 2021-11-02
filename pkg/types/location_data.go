package types

import (
	"time"

	"github.com/theotw/natssync/pkg/utils"
)

type LocationData struct {
	KeyID                string            `json:"keyID" bson:"keyID"`
	LocationID           string            `json:"locationID" bson:"locationID"`
	PublicKey            []byte            `json:"publicKey" bson:"publicKey"`
	PrivateKey           []byte            `json:"privateKey" bson:"privateKey"`
	Metadata             map[string]string `json:"metadata" bson:"metadata"`
	Created              time.Time         `json:"created" bson:"created"`
	LastModified         time.Time         `json:"lastModified" bson:"lastModified"`
	LastKeypairRotation  time.Time         `json:"lastKeypairRotation" bson:"lastKeypairRotation"`
	ForceKeypairRotation bool              `json:"forceKeypairRotation" bson:"forceKeypairRotation"`
}

func NewLocationData(
	locationID string,
	publicKey []byte,
	privateKey []byte,
	metadata map[string]string,
) (*LocationData, error) {

	// we want this to be a v1 UUID, this uuid will contain the creation timestamp
	// this will allow us to find the latest key
	newKeyID, err := utils.NewUUIDv1()
	if err != nil {
		return nil, err
	}

	data := &LocationData{
		KeyID:      newKeyID.String(),
		LocationID: locationID,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Metadata:   metadata,
	}

	return data.UpdateCreated(), nil
}

func (l *LocationData) SetForcedKeypairRotation() *LocationData {
	l.ForceKeypairRotation = true
	return l.UpdateLastModified()
}

func (l *LocationData) UnsetForcedKeyPairRotation() *LocationData {
	l.ForceKeypairRotation = false
	return l.UpdateLastModified()
}

func (l *LocationData) GetForceKeypairRotation() bool {
	return l.ForceKeypairRotation
}

func (l *LocationData) UpdateLastKeyPairRotation() *LocationData {
	l.LastKeypairRotation = time.Now()
	return l.UpdateLastModified()
}

func (l *LocationData) GetLastKeyPairRotation() time.Time {
	return l.LastKeypairRotation
}

func (l *LocationData) UpdateLastModified() *LocationData {
	l.LastModified = time.Now()
	return l
}

func (l *LocationData) GetLastModified() time.Time {
	return l.LastModified
}

func (l *LocationData) UpdateCreated() *LocationData {
	now := time.Now()
	l.Created = now
	l.LastKeypairRotation = now
	l.Created = now

	return l
}

func (l *LocationData) GetCreated() time.Time {
	return l.Created
}

func (l *LocationData) GetLocationID() string {
	return l.LocationID
}

func (l *LocationData) SetLocationID(locationID string) *LocationData {
	l.LocationID = locationID
	return l.UpdateLastModified()
}

func (l *LocationData) GetPublicKey() []byte {
	return l.PublicKey
}

func (l *LocationData) GetPrivateKey() []byte {
	return l.PrivateKey
}

func (l *LocationData) GetKeyPair() (publicKey []byte, privateKey []byte) {
	return l.GetPublicKey(), l.GetPrivateKey()
}

func (l *LocationData) SetKeyPair(publicKey, privateKey []byte) *LocationData {
	l.PublicKey = publicKey
	l.PrivateKey = privateKey
	return l.UpdateLastModified().UpdateLastKeyPairRotation()
}

func (l *LocationData) GetMetadata() map[string]string {
	return l.Metadata
}

func (l *LocationData) SetMetadata(metadata map[string]string) *LocationData {
	l.Metadata = metadata
	return l.UpdateLastModified()
}

func (l *LocationData) GetKeyID() string  {
	return l.KeyID
}

func (l *LocationData) SetKeyID(keyID string) error {
	if _,err := utils.ParseUUIDv1(keyID); err != nil {
		return err
	}
	l.KeyID = keyID
	return nil
}

func (l *LocationData) UnsetKeyID() *LocationData{
	l.KeyID = ""
	return l
}
