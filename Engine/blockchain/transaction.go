package blockchain

import (
	"bytes"
	"encoding/binary"
	"engine/database"
	"engine/utils"
	"errors"
	"time"

	"golang.org/x/crypto/sha3"
)

type Transaction struct {
	ID         HashBlock `json:"id"`          // Random hash created as the transaction ID.
	From       HashBlock `json:"from"`        // Account address "from"
	To         HashBlock `json:"to"`          // Account address "to"
	CreateTime uint64    `json:"create_time"` // Time when this transaction was created
	Ammount    float64   `json:"ammount"`     // Amount of the value being transferred
	Hash       HashBlock `json:"hash"`        // Hash of all the fields above
}

func (t *Transaction) GetHash() (result *HashBlock) {
	hash := sha3.New256()
	hash.Write(t.ID[:])
	hash.Write(t.From[:])
	hash.Write(t.To[:])
	hash.Write(utils.Uint64ToBytes(t.CreateTime))
	hash.Write(utils.Float64ToBytes(t.Ammount))
	digest := hash.Sum(nil)
	result = &HashBlock{}
	copy(result[:], digest)
	return result
}

/* Calculates and returns the hash string with the format 0xnnnnnnnn */
func (t *Transaction) ToHashString() (result string) {
	hash := t.GetHash()
	result = utils.EncodeHexString(hash[:])
	return result
}

/* Create new transaction */
func (a *Transaction) NewTransaction(from, to string, ammount float64) (result *Transaction, err error) {
	account := Account{}

	accountFrom, err := account.GetAccount(from)
	if err != nil {
		return nil, err
	}

	accountTo, err := account.GetAccount(to)
	if err != nil {
		return nil, err
	}

	if accountFrom.Equals(accountTo) {
		return nil, errors.New("attempting to transfer to the same account (from = to)")
	}

	if accountFrom.Balance < ammount {
		return nil, ErrInsufficientFunds
	}

	result = &Transaction{
		ID:         utils.NewRandomHash(),
		From:       accountFrom.Address,
		To:         accountTo.Address,
		CreateTime: uint64(time.Now().Unix()),
		Ammount:    ammount,
	}

	hash := result.GetHash()
	result.Hash.Set(hash)

	err = result.Persist()

	return result, err
}

func (a *Transaction) Persist() (err error) {

	db := &database.DatabaseFile{}
	err = db.Open(database.AccountsFileName)
	if err != nil {
		return err
	}
	defer db.Close()

	buff := &bytes.Buffer{}
	binary.Write(buff, binary.LittleEndian, *a)

	db.Write(buff.Bytes())

	return err
}
