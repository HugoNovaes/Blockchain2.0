package blockchain

import (
	"bytes"
	"engine/utils"

	"golang.org/x/crypto/sha3"
)

type HashBlock [32]byte

func (a *HashBlock) Compare(b *HashBlock) int {
	return bytes.Compare(a[:], b[:])
}

func (a *HashBlock) CompareBytes(b []byte) int {
	return bytes.Compare(a[:], b)
}

func (a *HashBlock) Equal(b *HashBlock) bool {
	return bytes.Equal(a[:], b[:])
}

func (a *HashBlock) Set(b *HashBlock) *HashBlock {
	a.SetBytes(b[:])
	return a
}

func (a *HashBlock) SetHexString(b string) (err error) {
	bytes, err := utils.DecodeHexString(b)
	if err != nil {
		return err
	}

	a.SetBytes(bytes)
	return nil
}

func (a *HashBlock) SetBytes(b []byte) *HashBlock {
	copy(a[:], b)
	return a
}

func (h *HashBlock) HashString(str string) *HashBlock {
	hash := sha3.New256()
	hash.Write([]byte(str))
	h.SetBytes(hash.Sum(nil))
	return h
}

func (h *HashBlock) Sha256Nonced(nonce *Nonce) *HashBlock {
	hash := sha3.New256()
	hash.Write(h[:])
	hash.Write(nonce.Bytes())
	h.SetBytes(hash.Sum(nil))
	return h
}

func (h *HashBlock) String() string {
	return utils.EncodeHexString(h[:])
}
