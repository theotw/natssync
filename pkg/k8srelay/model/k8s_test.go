/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package models

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
	"reflect"
	"testing"
)

func TestNewRelayKubeConfig(t *testing.T) {
	type args struct {
		routeID  string
		relayURL string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"happy", args{"1234", "https://localhost:8443"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRelayKubeConfig(tt.args.routeID, tt.args.relayURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRelayKubeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil {
				assert.Greater(t, len(got), 1)
				config, err := clientcmd.RESTConfigFromKubeConfig(got)
				if assert.Nil(t, err, "Should be a valid kube config") {
					token := config.BearerToken
					assert.Equal(t, tt.args.routeID, token)
				}

			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewRelayKubeConfig() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
