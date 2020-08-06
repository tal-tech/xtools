package aesutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func AesEncrypt(msg []byte, key string) (string, error) {

	k := []byte(key)
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	msg = PKCS7Padding(msg, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])

	cryted := make([]byte, len(msg))
	blockMode.CryptBlocks(cryted, msg)

	return base64.RawURLEncoding.EncodeToString(cryted), nil
}

func AesDecrypt(cryted string, key string) ([]byte, error) {

	crytedByte, err := base64.RawURLEncoding.DecodeString(cryted)
	if err != nil{
		return nil, err
	}
	k := []byte(key)

	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])

	orig := make([]byte, len(crytedByte))

	defer func(){
		if err := recover(); err != nil {
			return
		}
	}()

	blockMode.CryptBlocks(orig, crytedByte)

	orig = PKCS7UnPadding(orig)
	return orig, nil
}

func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

