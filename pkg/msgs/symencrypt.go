/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

func DoAesCBCEncrypt(src, key []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	src = addZeroPadding(src, bs)
	iv := make([]byte, bs)
	_, err = rand.Read(iv)
	if err != nil {
		return nil, err
	}

	if len(src)%bs != 0 {
		return nil, errors.New("pad failure")
	}
	cbcMode := cipher.NewCBCEncrypter(block, iv)
	out := make([]byte, len(src))
	cbcMode.CryptBlocks(out, src)
	ret := make([]byte, len(out)+len(iv))
	copy(ret, iv)
	copy(ret[bs:], out)
	return ret, nil
}

func DoAesCBCDecrypt(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	out := make([]byte, len(src)-bs)
	iv := src[:bs]

	if len(src)%bs != 0 {
		return nil, errors.New("not padded properly")
	}

	cbcMode := cipher.NewCBCDecrypter(block, iv)
	cbcMode.CryptBlocks(out, src[bs:])

	out = unPadTheZeros(out)
	return out, nil
}

func DoAesECBEncrypt(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	src = addZeroPadding(src, bs)
	if len(src)%bs != 0 {
		return nil, errors.New("pad failure")
	}
	out := make([]byte, len(src))
	dst := out
	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

func DoAesECBDecrypt(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(src))
	dst := out
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		return nil, errors.New("not padded properly")
	}
	for len(src) > 0 {
		block.Decrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	out = unPadTheZeros(out)
	return out, nil
}
