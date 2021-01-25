/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package integration

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/theotw/natssync/pkg"
	"net/http"
	"testing"
)

func TestServerAPI(t *testing.T) {
	t.Run("Test About", testAbout)
}
func testAbout(t *testing.T) {
	url := pkg.GetEnvWithDefaults("syncserver_url", "http://sync-server:8080")
	aboutURL := fmt.Sprintf("%s/event-bridge/1/about", url)
	resp, err := http.Get(aboutURL)
	if !assert.Nil(t, err) {
		t.Fatalf("Error not nil")
	}
	if !assert.NotNil(t, resp) {
		t.Fatalf("Resp is nil")
	}

	assert.Equal(t, 200, resp.StatusCode)
}
