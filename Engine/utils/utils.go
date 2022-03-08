package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/sha3"
)

func FileExists(filename string) bool {
	f, err := os.Open(filename)

	if os.IsNotExist(err) {
		return false
	}

	f.Close()

	return true
}

func FileIsEmpty(filename string) bool {
	f, err := os.Open(filename)

	if os.IsNotExist(err) {
		return true
	}
	defer f.Close()

	fs, _ := f.Stat()

	return fs.Size() == 0
}

func DecodeHexString(hexStr string) (bytes []byte, err error) {
	hexStr = strings.Replace(strings.ToUpper(hexStr), "0X", "", -1)
	bytes, err = hex.DecodeString(hexStr)
	return bytes, err
}

func EncodeHexString(bytes []byte) (result string) {

	if len(bytes) == 0 {
		result = "0x" + strings.Repeat("0", 64)
	} else {
		result = fmt.Sprintf("0x%x", bytes)
	}

	return result
}

func GenerateRSAKeyPair(keySize int) (publicKey []byte, privateKey []byte) {
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		log.Printf("GenerateRSAKeyPair(): %s\r\n", err.Error())
		return nil, nil
	}

	publicKey = x509.MarshalPKCS1PublicKey(key.Public().(*rsa.PublicKey))
	privateKey = x509.MarshalPKCS1PrivateKey(key)

	return publicKey, privateKey
}

func NewRandomHash() (result [32]byte) {

	var (
		guid       = uuid.New()
		hash       = sha3.New256()
		hashSeed   = time.Now().UnixNano()
		timeBuffer = make([]byte, 8)
	)

	binary.BigEndian.PutUint64(timeBuffer, uint64(hashSeed))
	hash.Write(timeBuffer)
	hash.Write(guid[:])

	copy(result[:], hash.Sum(nil))

	return result
}

func Uint64ToBytes(num uint64) (result []byte) {
	buff := bytes.Buffer{}
	binary.Write(&buff, binary.BigEndian, num)
	result = buff.Bytes()
	return result
}

func Float64ToBytes(num float64) (result []byte) {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, num)
	result = buf.Bytes()
	return result
}
