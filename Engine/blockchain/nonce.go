package blockchain

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"engine/utils"
	"errors"

	"golang.org/x/crypto/sha3"
)

type NonceBlock [16]byte

type Nonce struct {
	nonce NonceBlock
}

func (n *Nonce) Generate() *Nonce {
	rand.Read(n.nonce[:])
	return n
}

func (n *Nonce) Uint16() uint16 {
	return binary.BigEndian.Uint16(n.nonce[:])
}

func (n *Nonce) Uint32() uint32 {
	return binary.BigEndian.Uint32(n.nonce[:])
}

func (n *Nonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n.nonce[:])
}

func (n *Nonce) Bytes() []byte {
	return n.nonce[:]
}

func (n *Nonce) SetUint16(num uint16) *Nonce {
	binary.BigEndian.PutUint16(n.nonce[:], num)
	return n
}

func (n *Nonce) SetUint32(num uint32) *Nonce {
	binary.BigEndian.PutUint32(n.nonce[:], num)
	return n
}

func (n *Nonce) SetUint64(num uint64) *Nonce {
	binary.BigEndian.PutUint64(n.nonce[:], num)
	return n
}

func (n *Nonce) Get() *NonceBlock {
	return &n.nonce
}

func (n *Nonce) Set(b *Nonce) {
	copy(n.nonce[:], b.nonce[:])
}

func (n *Nonce) ToString() string {
	return "0x" + hex.EncodeToString(n.Bytes())
}

func (n *Nonce) Sha256() (result *Nonce) {
	hash := sha3.New256()
	hash.Write(n.Bytes())
	result = &Nonce{}
	result.SetBytes(hash.Sum(nil))
	return result

}

func (n *Nonce) FromString(hexStr string) error {

	b, err := utils.DecodeHexString(hexStr)
	if err != nil {
		return err
	}
	n.SetBytes(b)

	return nil
}

func (n *Nonce) SetBytes(b []byte) (*Nonce, error) {

	if len(b) == 0 || len(b) > len(n.nonce) {
		return nil, errors.New("invalid nonce size")
	}

	copy(n.nonce[:], b)

	return n, nil
}

func SwapBytes(b []byte) []byte {
	bufferLength := len(b)
	buffer := make([]byte, bufferLength)

	for i := 0; i < bufferLength; i++ {
		buffer[i] = b[bufferLength-i-1]
	}

	return buffer
}
