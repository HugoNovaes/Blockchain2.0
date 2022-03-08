package blockchain

import (
	"encoding/json"
	"engine/database"
	"engine/utils"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	GenesisFileName = "./db/genesis.json"

	mySignature = "HSN Blockchain - Developed by Hugo de Souza Novaes - hnovaes@yahoo.com"
	Coinbase    = "0x1c6ab7bbf2e4ca7c68a2f455c6e3dcc10ad5b5a5"
	Version     = "1.0.0"
)

type (
	Block struct {
		Id         uint64     `json:"id"`
		Parent     HashBlock  `json:"parent"`
		Hash       HashBlock  `json:"hash"`
		Nonce      NonceBlock `json:"nonce"`
		Merkle     HashBlock  `json:"merkle"`
		Difficulty uint64     `json:"difficulty"`
		Time       uint64     `json:"time"`
		Version    uint16     `json:"version"`
		Coinbase   HashBlock  `json:"coinbase"`
	}

	Blockchain struct {
		mutex                sync.Mutex
		creatingGenesisBlock bool
		Blocks               []Block
	}
)

func (b *Blockchain) CurrentBlock() *Block {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.creatingGenesisBlock {
		return nil
	}

	b.checkAndLoadBlocks()
	lenBlocks := len(b.Blocks)
	return &b.Blocks[lenBlocks-1]
}

func (b *Blockchain) NewHash(newHash *HashBlock, nonce *Nonce) (*HashBlock, bool) {

	block := b.CurrentBlock()

	cmp := newHash.Compare(&block.Hash)
	accepted := cmp < 0

	if !accepted {
		return nil, false
	}

	newBlock := b.NewBlock(newHash, nonce)

	if newBlock == nil {
		return nil, false
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.Blocks = append(b.Blocks, *newBlock)

	b.Persist(false)

	return newHash, accepted
}

func (b *Blockchain) Persist(isGenesis bool) (err error) {
	dat := database.BlockDB{}
	dat.Open()
	defer dat.Close()

	data, err := json.Marshal(b.Blocks[len(b.Blocks)-1])
	if err != nil {
		return err
	}

	return dat.Add(data)
}

func (b *Blockchain) NewBlock(newHash *HashBlock, newNonce *Nonce) *Block {

	var blockId uint64
	var parentHash HashBlock

	if b.creatingGenesisBlock {
		return &Block{
			Id:         0,
			Parent:     parentHash,
			Hash:       *newHash,
			Nonce:      newNonce.nonce,
			Time:       uint64(time.Now().Unix()),
			Difficulty: 1,
		}
	}

	var lastBlock = b.CurrentBlock()
	blockId = lastBlock.Id + 1
	parentHash = lastBlock.Hash

	newBlock := &Block{
		Id:         blockId,
		Parent:     parentHash,
		Hash:       *newHash,
		Nonce:      newNonce.nonce,
		Time:       uint64(time.Now().Unix()),
		Difficulty: lastBlock.Difficulty,
		Coinbase:   lastBlock.Coinbase,
		Version:    lastBlock.Version,
	}

	return newBlock
}

func (b *Blockchain) createGenesisHash(threaId int, wg *sync.WaitGroup, cbTestNewGenesis func(*Block) bool) {

	var (
		maxHash    = &HashBlock{}
		oldMaxHash = &HashBlock{}
		maxNonce   = &Nonce{}
		newHash    = &HashBlock{}
		newNonce   = &Nonce{}
		maxTime    = time.Now().Format(time.RFC3339)
	)

	defer wg.Done()

	newHash.HashString(mySignature)
	newHash[0] = 0
	newHash[1] = 0
	maxHash.Set(newHash)

	startTime := time.Now()
	log.SetPrefix("\r")

	for checkpoint := uint64(0); ; checkpoint++ {

		newNonce.Generate()
		newHash.Sha256Nonced(newNonce)

		if newHash.Compare(maxHash) > 0 && newHash[0] == 0 {
			maxHash.Set(newHash)
			maxNonce.Set(newNonce)
			maxTime = time.Now().Format(time.RFC3339)

			if !oldMaxHash.Equal(maxHash) {
				genesis := b.NewBlock(maxHash, maxNonce)
				accepted := cbTestNewGenesis(genesis)
				if accepted {
					log.SetPrefix("\r")
					log.Printf("H:%x N:%x T:%s\r\n", maxHash[:], maxNonce.Bytes(), maxTime)
				}

				oldMaxHash.Set(maxHash)
			}
		}

		if checkpoint%500000 == 0 {
			elspsedTime := time.Since(startTime)
			fmt.Printf("\r%s", elspsedTime)

			if elspsedTime.Minutes() >= 1 {
				break
			}
		}
	}
}

func (b *Blockchain) createGenesisBlock() error {

	funcCreatingGenesisBlock := func(bc *Blockchain, status bool) {
		bc.creatingGenesisBlock = status
	}

	funcCreatingGenesisBlock(b, true)
	defer funcCreatingGenesisBlock(b, false)

	fmt.Println("Generating Genesis Block...")
	fmt.Println("This task will spend some time while calculating the best hash for the Genesis Block.")

	genesisFile, err := os.OpenFile(GenesisFileName, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	defer genesisFile.Close()

	wg := &sync.WaitGroup{}
	wg.Add(runtime.NumCPU())
	var mutex = &sync.Mutex{}
	var targetGenesisBlock *Block

	for i := 0; i < runtime.NumCPU(); i++ {

		go b.createGenesisHash(i, wg, func(genesis *Block) bool {
			mutex.Lock()
			defer mutex.Unlock()

			if targetGenesisBlock == nil {
				targetGenesisBlock = genesis
				return true
			}

			hashNewBlock := genesis.Hash
			hashTargetBlock := &targetGenesisBlock.Hash

			var bytesNewBlockHash uint32 = uint32(hashNewBlock[2])<<24 | uint32(hashNewBlock[3])<<16 | uint32(hashNewBlock[4])<<8 | uint32(hashNewBlock[5])
			var bytesTargetBlock uint32 = uint32(hashTargetBlock[2])<<24 | uint32(hashTargetBlock[3])<<16 | uint32(hashTargetBlock[4])<<8 | uint32(hashTargetBlock[5])

			cmp := hashNewBlock.Compare(hashTargetBlock)
			if cmp <= 0 && bytesNewBlockHash < bytesTargetBlock {
				return false
			}

			targetGenesisBlock = genesis

			return true
		})
	}
	wg.Wait()

	jstr, _ := json.Marshal(targetGenesisBlock)
	log.Println("Genesis block created:")
	fmt.Println(string(jstr))

	err = os.WriteFile(GenesisFileName, jstr, 0664)
	if err != nil {
		log.Printf("Error creating a copy of %s: %s\r\n", GenesisFileName, err.Error())
	}

	b.Blocks = make([]Block, 1)
	b.Blocks[0] = *targetGenesisBlock

	err = b.Persist(true)
	return err
}

func (b *Blockchain) LoadBlockchainDatabase() {
	if !utils.FileExists(GenesisFileName) || utils.FileIsEmpty(GenesisFileName) {
		if err := b.createGenesisBlock(); err != nil {
			log.Panicf("Cannot create file %s\r\n", GenesisFileName)
		}
	}

	db := database.BlockDB{}
	b.Blocks = make([]Block, 0)

	err := db.Open()
	if err != nil {
		log.Panicf("Cannot open database file %s\r\n", err)
	}

	db.LoadData(func(data []byte) {
		block := Block{}
		err = json.Unmarshal(data, &block)
		b.Blocks = append(b.Blocks, block)
	})
}

func (b *Blockchain) checkAndLoadBlocks() {

	if len(b.Blocks) > 0 {
		return
	}

	b.LoadBlockchainDatabase()
}
