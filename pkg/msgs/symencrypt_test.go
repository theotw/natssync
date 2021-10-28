/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theotw/natssync/pkg/msgs"
	_ "github.com/theotw/natssync/tests/unit"
)

func Test_AesEnryption(t *testing.T) {
	testData := []string{"hello", "hello world", "super long weird data with \t an                        d stuff     "}
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.Nil(t, err)
	for i, plain1 := range testData {
		name := fmt.Sprintf("Name %d", i)
		t.Run(name, func(t *testing.T) {
			cipher, err := msgs.DoAesECBEncrypt([]byte(plain1), key)
			assert.Nil(t, err)
			if err == nil {
				plain2, err := msgs.DoAesECBDecrypt(cipher, key)
				assert.Nil(t, err)
				if err == nil {
					plainText2 := string(plain2)
					assert.Equal(t, plain1, plainText2)
				}
			}
		})
	}
}
func Test_AesCBCEnryption(t *testing.T) {
	testData := []string{"hello", "hello world", "super long weird data with \t an                        d stuff     "}
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.Nil(t, err)
	for i, plain1 := range testData {
		name := fmt.Sprintf("Name %d", i)
		t.Run(name, func(t *testing.T) {
			cipher, err := msgs.DoAesCBCEncrypt([]byte(plain1), key)
			assert.Nil(t, err)
			if err == nil {
				plain2, err := msgs.DoAesCBCDecrypt(cipher, key)
				assert.Nil(t, err)
				if err == nil {
					plainText2 := string(plain2)
					assert.Equal(t, plain1, plainText2)
				}
			}
		})
	}
}
