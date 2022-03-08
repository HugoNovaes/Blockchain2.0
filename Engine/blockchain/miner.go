package blockchain

import (
	"log"

	"golang.org/x/crypto/sha3"
)

type Miner struct {
	Blockchain    *Blockchain
	ThreadId      int
	foundCallBack func(*HashBlock, *Nonce)
}

func (m *Miner) StartMiner(bc *Blockchain, foundHashCallBack func(*HashBlock, *Nonce)) {
	m.Blockchain = bc
	m.foundCallBack = foundHashCallBack
	go m.RunMiner()
}

func (m *Miner) Difficulty() uint64 {
	return m.Blockchain.CurrentBlock().Difficulty
}

func (m *Miner) RunMiner() {

	block := m.Blockchain.CurrentBlock()
	targetHash := &block.Hash

	nonce := &Nonce{}
	nonce.SetBytes(block.Nonce[:])
	nonce.SetUint64(nonce.Uint64() << (m.ThreadId - 1) << 8)

	for {
		hashToVerify := GenerateHash(nonce)
		isItGoodOne := m.VerifyHash(hashToVerify, targetHash)

		if isItGoodOne {
			newHash, accepted := m.Blockchain.NewHash(hashToVerify, nonce)

			if accepted {

				go func(b *Blockchain, thid int, diff uint64, hsh []byte, n *Nonce) {
					bkp := log.Prefix()
					log.SetPrefix("\r\n")
					log.Printf("[B:%d H:%x N:%032x D:%d T:%d]", b.CurrentBlock().Id, hsh, n.Bytes(), diff, thid)
					log.SetPrefix(bkp)
				}(m.Blockchain, m.ThreadId, block.Difficulty, hashToVerify[:], nonce)

				targetHash = newHash
				block = m.Blockchain.CurrentBlock()
				if m.foundCallBack != nil {
					m.foundCallBack(hashToVerify, nonce)
				}
			}
		}
		nonce.Generate()
	}
}

func GenerateHash(nonce *Nonce) (result *HashBlock) {
	hash := sha3.New256()
	hash.Write(nonce.Bytes())
	result = &HashBlock{}
	result.SetBytes(hash.Sum(nil))
	return result
}

func (m *Miner) VerifyHash(hashToVerify *HashBlock, targetHash *HashBlock) bool {

	diff := m.Difficulty()

	for i := uint64(0); i < diff; i++ {
		if hashToVerify[i] != 0 {
			return false
		}
	}

	return hashToVerify.Compare(targetHash) < 0
}
