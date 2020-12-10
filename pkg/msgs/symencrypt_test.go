package msgs

import (
	"crypto/des"
	"crypto/rand"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DesEnryption(t *testing.T) {
	testData := []string{"hello", "hello world", "super long weird data with \t an                        d stuff     "}
	key := make([]byte, des.BlockSize)
	rand.Read(key)
	for i, plain1 := range testData {
		name := fmt.Sprintf("Name %d", i)
		t.Run(name, func(t *testing.T) {
			cipher, err := DoDesEncrypt([]byte(plain1), key)
			assert.Nil(t, err)
			if err == nil {
				plain2, err := DoDesDecrypt(cipher, key)
				assert.Nil(t, err)
				if err == nil {
					plainText2 := string(plain2)
					assert.Equal(t, plain1, plainText2)
				}
			}
		})
	}
}
func Test_AesEnryption(t *testing.T) {
	testData := []string{"hello", "hello world", "super long weird data with \t an                        d stuff     "}
	key := make([]byte, 32)
	rand.Read(key)
	for i, plain1 := range testData {
		name := fmt.Sprintf("Name %d", i)
		t.Run(name, func(t *testing.T) {
			cipher, err := DoAesEncrypt([]byte(plain1), key)
			assert.Nil(t, err)
			if err == nil {
				plain2, err := DoAesDecrypt(cipher, key)
				assert.Nil(t, err)
				if err == nil {
					plainText2 := string(plain2)
					assert.Equal(t, plain1, plainText2)
				}
			}
		})
	}
}
