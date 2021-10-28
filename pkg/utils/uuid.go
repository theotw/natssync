package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	uuidv1VersionString = "VERSION_1"
)

type UUIDv1 struct {
	uuid uuid.UUID
}

func NewUUIDv1() (*UUIDv1, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	return &UUIDv1{
		uuid: id,
	}, nil
}

func ParseUUIDv1(uuidString string) (*UUIDv1, error) {
	id, err := uuid.Parse(uuidString)
	if err != nil {
		return nil, err
	}
	if id.Version().String() != uuidv1VersionString {
		return nil, fmt.Errorf("invalid UUID expected %s got %s", uuidv1VersionString, id.Version().String())
	}
	return &UUIDv1{
		uuid: id,
	}, nil

}

func (u *UUIDv1) GetCreationTime() time.Time {
	sec, nsec := u.uuid.Time().UnixTime()
	return time.Unix(sec, nsec).UTC()
}

func (u *UUIDv1) String() string {
	return u.uuid.String()
}

type UUIDv4 struct {
	uuid uuid.UUID
}

func NewUUIDv4() *UUIDv4 {
	id := uuid.New()
	return &UUIDv4{
		uuid: id,
	}
}

func (u *UUIDv4) String() string {
	return u.uuid.String()
}
