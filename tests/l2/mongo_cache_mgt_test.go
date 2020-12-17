/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package l2

import (
	"github.com/stretchr/testify/assert"
	"github.com/theotw/natssync/pkg/cloudserver"
	"testing"
)

func TestMongo(t *testing.T){
	err:=cloudserver.DoMongoStuff()
	assert.Nil(t,err)
}
