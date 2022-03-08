package blockchain

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Node struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

var userSelectedPort = 0

func (n *Node) ToString() string {
	if userSelectedPort > 0 {
		return fmt.Sprintf(":%d", userSelectedPort)
	}

	return fmt.Sprintf(":%d", n.Port)
}

type Config struct {
	ParentNode *Node   `json:"parent_node"`
	ThisNode   *Node   `json:"this_node"`
	ChildNodes []*Node `json:"child_nodes"`
}

type BlockchainNode struct {
	Configuration     *Config
	Listener          net.Listener
	ClientConnections []net.Conn
}

func (n *BlockchainNode) UsePort(portToUse int) {
	userSelectedPort = portToUse
}

func (n *BlockchainNode) loadConfig() (err error) {
	data, err := os.ReadFile(ConfigFileName)
	if err != nil {
		return err
	}

	n.Configuration = &Config{}
	err = json.Unmarshal(data, n.Configuration)

	if userSelectedPort > 0 {
		n.Configuration.ThisNode.Port = userSelectedPort
	}

	return err
}

func (n *BlockchainNode) AddClient(client net.Conn) {
	n.ClientConnections = append(n.ClientConnections, client)
}

func (n *BlockchainNode) AcceptConnection(client net.Conn) {
	remaddr := client.RemoteAddr()
	log.Printf("New connection %s %s\r\n", remaddr.Network(), remaddr.String())
	defer client.Close()

	n.AddClient(client)
	n.Send(client, CMD_WELCOME, []byte("Welcome!"), 0, "Success")

	strBuffer := ""
	log.SetPrefix("\r")

	for {
		var buffer [8192]byte
		numBytesRead, err := client.Read(buffer[:])
		if err != nil {
			log.Println(err.Error())
			break
		}

		if numBytesRead <= 0 {
			log.Printf("%s disconnected.\r\n", client.RemoteAddr())
			break
		}

		strBuffer += string(buffer[:numBytesRead])
		for {
			packEndPos := strings.IndexByte(strBuffer, '}')
			if packEndPos <= 0 {
				break
			}

			strPack := strBuffer[0 : packEndPos+1]
			pack := &DataPack{}
			err := json.Unmarshal([]byte(strPack), pack)

			if err != nil {
				n.Send(client, CMD_ERROR, []byte("Error parsing json data"), ERR_JSON_PARSING, err.Error())
				log.Printf("Error from %s: %s\r\n", client.RemoteAddr(), err.Error())
			}

			strBuffer = strBuffer[packEndPos+1:]
		}
	}

}

func (n *BlockchainNode) RunServerNode() {
	var err error
	var thisNode = n.Configuration.ThisNode
	log.SetPrefix("\r")
	log.Println("Starting node server...")

	n.Listener, err = net.Listen("tcp4", thisNode.ToString())
	if err != nil {
		log.Fatalf("Couldn't start Node Server on port %d\r\n", thisNode.Port)
	}

	log.Printf("Node server listening on %s\r\n", thisNode.ToString())

	for {
		client, err := n.Listener.Accept()
		if err != nil {
			log.Printf("New connection error: %s\r\n", err.Error())
			continue
		}

		go n.AcceptConnection(client)
	}
}

func (n *BlockchainNode) StartListener() {
	err := n.loadConfig()

	if err != nil {
		log.Fatalf("Error loading config.json: %s\r\n", err.Error())
	}
	go n.RunServerNode()
}
