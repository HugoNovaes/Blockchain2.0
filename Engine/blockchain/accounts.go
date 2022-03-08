package blockchain

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"engine/database"
	"engine/utils"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/sha3"
)

type (
	Account struct {
		Label      [64]byte   `json:"label"`
		Address    HashBlock  `json:"address"`
		CreateTime uint64     `json:"create_time"`
		Balance    float64    `json:"balance"`
		PublicKey  [8192]byte `json:"public_key"`
		PrivateKey [8192]byte `json:"private_key"`
	}
)

var AccountCache []Account = make([]Account, 0)

func (a *Account) Equals(b *Account) bool {
	if a == nil && b == nil {
		return true
	}

	cmpLabel := bytes.Compare(a.Label[:], b.Label[:])
	cmpAddrs := bytes.Compare(a.Address[:], b.Address[:])

	if cmpLabel == 0 && cmpAddrs == 0 {
		return true
	}

	return false
}

func NewAccount(label string) (result *Account, err error) {
	if len(label) == 0 || len(label) > 64 {
		return nil, errors.New("account label has invalid length. Maximum length is 64 characters")
	}

	newAddress := utils.NewRandomHash()
	publicKey, privateKey := utils.GenerateRSAKeyPair(2048)

	result = &Account{
		Address:    newAddress,
		CreateTime: uint64(time.Now().Unix()),
	}

	copy(result.Label[:], []byte(label))
	copy(result.PublicKey[:], publicKey)
	copy(result.PrivateKey[:], privateKey)

	err = result.Persist()
	if err != nil {
		return nil, err
	}

	return result, err
}

func (a *Account) SignWithPrivateKey(dataToSign []byte, accAddress string) (result []byte, err error) {

	db := &database.DatabaseFile{}
	err = db.Open(database.AccountsFileName)
	defer db.Close()
	if err != nil {
		return nil, err
	}

	bytesAccAddress, err := utils.DecodeHexString(accAddress)

	prvKey, err := db.FindData(func(data []byte) []byte {

		decData, err := utils.AESGCMDecrypt(data)
		if err != nil {
			fmt.Printf("%s\r\n", err.Error())
			return nil
		}

		acc := &Account{}
		err = json.Unmarshal(decData, acc)

		if err != nil {
			fmt.Println(err.Error())
			return nil
		}

		if acc.Address.CompareBytes(bytesAccAddress) == 0 {
			return decData
		}

		return nil
	})

	rsaPrvKey, err := x509.ParsePKCS1PrivateKey(prvKey)
	if err != nil {
		return nil, err
	}

	hash := sha3.New256()
	hash.Write(dataToSign)

	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrvKey, crypto.SHA256, hash.Sum(nil))
	if err != nil {
		return nil, err
	}

	return signature, err
}

func (a *Account) LoadAccountsDatabase() (result []Account) {
	db := database.DatabaseFile{}
	db.Open(database.AccountsFileName)
	defer db.Close()

	result = make([]Account, 0)
	db.ForEach(func(data []byte) {
		account := Account{}

		buff := &bytes.Buffer{}
		buff.Write(data)
		err := binary.Read(buff, binary.LittleEndian, &account)
		if err != nil {
			fmt.Println(err.Error())
		}
		result = append(result, account)
	})
	return result
}

func (a *Account) Persist() (err error) {
	if len(AccountCache) == 0 {
		AccountCache = a.LoadAccountsDatabase()
	}

	db := &database.DatabaseFile{}
	err = db.Open(database.AccountsFileName)
	defer db.Close()

	if err != nil && !errors.Is(err, database.ErrEmpty) {
		return err
	}
	buff := &bytes.Buffer{}
	err = binary.Write(buff, binary.LittleEndian, a)

	hashString := utils.EncodeHexString(a.Address[:])
	account, err := a.GetAccount(hashString)
	if errors.Is(err, ErrAccountNotFound) {
		AccountCache = append(AccountCache, *a)
		return db.Write(buff.Bytes())
	}

	*account = *a
	db.Update(buff.Bytes(), func(data []byte) bool {
		acc := &Account{}
		buff := &bytes.Buffer{}
		buff.Write(data)
		binary.Read(buff, binary.LittleEndian, acc)

		if bytes.Compare(acc.Label[:], a.Label[:]) == 0 {
			return true
		}
		return false
	})

	return err
}

func (a *Account) CheckIntegrity() (err error) {
	return err
}

func (a *Account) GetAccount(address string) (result *Account, err error) {
	if len(AccountCache) == 0 {
		AccountCache = a.LoadAccountsDatabase()
	}

	hash, err := utils.DecodeHexString(address)
	if err != nil {
		return nil, err
	}

	for _, account := range AccountCache {
		if account.Address.Compare((*HashBlock)(hash)) == 0 {
			return &account, nil
		}
	}

	return result, ErrAccountNotFound
}

func (a *Account) ListAll() {
	if len(AccountCache) == 0 {
		AccountCache = a.LoadAccountsDatabase()
	}

	for _, account := range AccountCache {
		fmt.Printf("Address: 0x%x Balance: %0.8f\r\n", account.Address, account.Balance)
	}
}

func Airdrop(to string, ammount float64) (result *Account, err error) {
	result, err = (&Account{}).GetAccount(to)
	if err != nil {
		return nil, err
	}

	result.Balance += ammount
	result.Persist()

	return result, err
}
