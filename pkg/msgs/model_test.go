/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs_test

import (
	"reflect"
	"testing"

	"github.com/theotw/natssync/pkg/msgs"
)

func TestParseSubject(t *testing.T) {

	happyNoParam := msgs.MakeMessageSubject("1", "")
	happyParam := msgs.MakeMessageSubject("1", "test")
	badSub := "bob"
	tests := []struct {
		name    string
		subject string
		want    *msgs.ParsedSubject
		wantErr bool
	}{
		{"happy no args", happyNoParam, &msgs.ParsedSubject{AppData: []string{}, OriginalSubject: happyNoParam, LocationID: "1"}, false},
		{"happy no args", happyParam, &msgs.ParsedSubject{AppData: []string{"test"}, OriginalSubject: happyParam, LocationID: "1"}, false},
		{"bad subject", badSub, &msgs.ParsedSubject{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := msgs.ParseSubject(tt.subject)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSubject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSubject() got = %v, want %v", got, tt.want)
			}
		})
	}
}
