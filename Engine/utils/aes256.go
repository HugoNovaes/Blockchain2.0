package utils

import (
	"crypto/aes"
	"crypto/cipher"
)

var (
	internalKey = []byte{
		0x41, 0x1F, 0x59, 0xE5, 0xD4, 0x76, 0x6E, 0x1B,
		0xBD, 0x6E, 0x5F, 0xBF, 0x73, 0xCB, 0x83, 0xA3,
		0xC2, 0x1B, 0x22, 0x74, 0xE5, 0xFD, 0x4E, 0x6C,
		0x13, 0x97, 0x2E, 0x0F, 0x9F, 0x67, 0x99, 0x23}
)

func AESGCMEncrypt(plaintext []byte) ([]byte, error) {

	block, err := aes.NewCipher(internalKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	copy(nonce, internalKey[:12])

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nil
}

func AESGCMDecrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(internalKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	copy(nonce, internalKey[:12])

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
