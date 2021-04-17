package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
)

type cipherParams struct {
	Ciphertext string `json:"ciphertext"`
	IV         string `json:"iv"`
}

func pad(b []byte) []byte {
	padSize := aes.BlockSize - (len(b) % aes.BlockSize)
	pad := bytes.Repeat([]byte{byte(padSize)}, padSize)
	return append(b, pad...)
}

func unpad(b []byte) []byte {
	padSize := int(b[len(b)-1])
	if padSize > aes.BlockSize || padSize > len(b) {
		return b
	}
	return b[:len(b)-padSize]
}

func encrypt(securityToken, input, ivString string) (string, error) {
	plainText := pad([]byte(input))
	secret, _ := base64.StdEncoding.DecodeString(securityToken)
	iv, _ := hex.DecodeString(ivString)
	block, _ := aes.NewCipher(secret)
	mode := cipher.NewCBCEncrypter(block, iv[:16])
	encrypted := make([]byte, len(plainText))
	mode.CryptBlocks(encrypted, plainText)
	cipherText := hex.EncodeToString(encrypted)
	params := cipherParams{
		Ciphertext: cipherText,
		IV:         ivString,
	}
	enc, _ := json.Marshal(&params)
	data := base64.StdEncoding.EncodeToString(enc)
	return data, nil
}

func decodeCipherParams(data string) (*cipherParams, error) {
	enc, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	var params cipherParams
	err = json.Unmarshal(enc, &params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func decrypt(securityToken string, data string) (string, error) {
	secret, err := base64.StdEncoding.DecodeString(securityToken)
	if err != nil {
		return "", err
	}
	params, err := decodeCipherParams(data)
	if err != nil {
		return "", err
	}
	iv, err := hex.DecodeString(params.IV)
	if err != nil {
		return "", err
	}
	cipherText, err := hex.DecodeString(params.Ciphertext)
	if err != nil {
		return "", err
	}
	c, err := aes.NewCipher(secret)
	if err != nil {
		return "", err
	}
	blockSize := aes.BlockSize
	mode := cipher.NewCBCDecrypter(c, iv[:blockSize])

	plainText := make([]byte, len(cipherText))
	mode.CryptBlocks(plainText, cipherText)
	return string(unpad(plainText)), nil
}

func generateRandomKey(length int) ([]byte, error) {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	return k, nil
}
