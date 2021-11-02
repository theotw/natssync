package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theotw/natssync/pkg/utils"
)

func TestParseUUIDv1(t *testing.T) {
	id, err := utils.NewUUIDv1()
	assert.Nil(t, err)
	idString := id.String()
	parsedId, err := utils.ParseUUIDv1(idString)
	assert.Nil(t, err)
	assert.Equal(t, id.String(), parsedId.String())
}
