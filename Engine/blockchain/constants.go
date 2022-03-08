package blockchain

import "errors"

const (
	ConfigFileName = "./nodeconfig.json"
)

var (
	ErrAccountNotFound   = errors.New("account does not exist")
	ErrInsufficientFunds = errors.New("insufficient funds to transfer")
	ErrNoTransactions    = errors.New("transactions list is empty")
)
