/*
 *  Copyright 2022 VMware, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *  http://www.apache.org/licenses/LICENSE-2.0
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

// Decrypt uses the given salt and encryptionKey to decrypt the given encrypted
// string using the same method as that used by the Cloud Controller.
func Decrypt(data string, salt string, encryptionKey string) (string, error) {
	key, iv := getKeyAndIV([]byte(encryptionKey), []byte(salt))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	encryptedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	decryptedData := encryptedData
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decryptedData, encryptedData)
	result, err := pkcs7Unpad(decryptedData, 16)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// Encrypt uses the given salt and encryptionKey to encrypt the given plaintext
// string using the same method as that used by the Cloud Controller.
func Encrypt(data string, salt string, encryptionKey string) (string, error) {
	key, iv := getKeyAndIV([]byte(encryptionKey), []byte(salt))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decryptedData, err := pkcs7Pad([]byte(data), 16)
	if err != nil {
		return "", err
	}

	encryptedData := decryptedData

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptedData, decryptedData)

	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

func getKeyAndIV(encryptionKey []byte, saltBytes []byte) ([]byte, []byte) {
	var (
		key []byte
		iv  []byte
	)
	// In older versions of PCF, a unusual method was used to calculate the salt
	// and encrypt the text. Thus, we have an unusual method of getting the key
	// and iv when the salt is short.
	if len(saltBytes) == 8 {
		key, iv = extractOpenSSLCreds(encryptionKey, saltBytes)
	} else {
		key = pbkdf2.Key(encryptionKey, saltBytes, 2048, 16, sha256.New)
		iv = saltBytes
	}

	return key, iv
}

func extractOpenSSLCreds(password, salt []byte) ([]byte, []byte) {
	prev := []byte{}
	m0 := hash(prev, password, salt)
	for i := 1; i < 2048; i++ {
		m0 = md5sum(m0)
	}
	m1 := hash(m0, password, salt)
	for i := 1; i < 2048; i++ {
		m1 = md5sum(m1)
	}

	return m0, m1
}

func hash(prev, password, salt []byte) []byte {
	a := make([]byte, len(prev)+len(password)+len(salt))
	copy(a, prev)
	copy(a[len(prev):], password)
	copy(a[len(prev)+len(password):], salt)
	return md5sum(a)
}

func md5sum(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

// pkcs7Unpad returns slice of the original data without padding.
func pkcs7Unpad(data []byte, blocklen int) ([]byte, error) {
	if blocklen <= 0 {
		return nil, fmt.Errorf("invalid blocklen %d", blocklen)
	}
	if len(data)%blocklen != 0 || len(data) == 0 {
		return nil, fmt.Errorf("invalid data len %d", len(data))
	}
	padlen := int(data[len(data)-1])
	if padlen > blocklen || padlen == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	pad := data[len(data)-padlen:]
	for i := 0; i < padlen; i++ {
		if pad[i] != byte(padlen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return data[:len(data)-padlen], nil
}

func pkcs7Pad(data []byte, blockSize int) ([]byte, error) {
	bufLen := len(data)
	padLen := blockSize - (bufLen % blockSize)
	padText := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padText...), nil
}
