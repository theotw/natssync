/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"reflect"
	"testing"
)

func TestParseSubject(t *testing.T) {

	happyNoParam := MakeMessageSubject("1", "")
	happyParam := MakeMessageSubject("1", "test")
	badSub := "bob"
	tests := []struct {
		name    string
		subject string
		want    *ParsedSubject
		wantErr bool
	}{
		{"happy no args", happyNoParam, &ParsedSubject{AppData: []string{}, OriginalSubject: happyNoParam, LocationID: "1"}, false},
		{"happy no args", happyParam, &ParsedSubject{AppData: []string{"test"}, OriginalSubject: happyParam, LocationID: "1"}, false},
		{"bad subject", badSub, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSubject(tt.subject)
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
