package blockchain

import (
	"encoding/json"
	"net"
)

const (
	CMD_WELCOME = 0x01
	CMD_ERROR   = 0x02
)

const (
	ERR_JSON_PARSING = 0x01
)

type DataPack struct {
	Command   uint8  `json:"c"`
	ErrorCode uint8  `json:"ec"`
	ErrorMsg  string `json:"em"`
	Data      []byte `json:"d"`
}

func (n *BlockchainNode) Send(conn net.Conn, cmd uint8, data []byte, errCode uint8, errMsg string) (err error) {

	pack := DataPack{
		Command:   cmd,
		ErrorCode: errCode,
		ErrorMsg:  errMsg,
		Data:      data,
	}

	bytes, _ := json.Marshal(pack)

	_, err = conn.Write(bytes)
	return err
}
