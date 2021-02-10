/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package l2

import (
	"github.com/stretchr/testify/assert"
	"github.com/theotw/natssync/pkg/cloudserver"
	"testing"
)

func TestMongo(t *testing.T) {
	err := cloudserver.DoMongoStuff()
	assert.Nil(t, err)
}
