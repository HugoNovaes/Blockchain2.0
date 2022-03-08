package main

import (
	"engine/blockchain"
	"engine/webserver"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type (
	Parameter struct {
		Value       string
		Required    bool
		Description string
	}

	Command struct {
		Description []string
		Parameters  map[string]*Parameter
		Func        func(c *Command)
	}
)

var (
	Commands map[string]*Command
)

func buildCommandList() {

	if len(Commands) > 0 {
		return
	}

	Commands = map[string]*Command{
		"?": {
			Description: []string{"Display the list of arguments"},
			Func:        displayHelp,
			Parameters:  map[string]*Parameter{},
		},

		"help": {
			Description: []string{"Display the list of arguments"},
			Func:        displayHelp,
			Parameters:  map[string]*Parameter{},
		},

		"airdrop": {
			Description: []string{"Blockchain deposit \"ammount\" of coins into the \"to\" account"},
			Func:        doAirdrop,
			Parameters: map[string]*Parameter{
				"to":      {Required: true, Description: "Destination of the ammount to be credited"},
				"ammount": {Required: true, Description: "the ammount to be transferred. Must be > 0 and <= 10000"},
			},
		},
		"send": {
			Description: []string{"Transfer coins from an account to a destination account."},
			Func:        doSend,
			Parameters: map[string]*Parameter{
				"from":    {Required: true, Description: "The account to be debited"},
				"to":      {Required: true, Description: "The account to be credited"},
				"ammount": {Required: true, Description: "The ammount to transfer. Must be > 0 and <= 10000"},
			},
		},
		"startminer": {
			Description: []string{"Start the blockchain miner engine."},
			Func:        doStartMiner,
			Parameters: map[string]*Parameter{
				"threads":   {Required: false, Description: "Number of threads to use. Default is the number of CPU Cores. Value must be >= 1 and limited to the SO capacity."},
				"benchmark": {Required: false, Description: "Start miner on benchmark mode. Value must be 'yes' or 'no'"},
			},
		},
		"accounts": {
			Description: []string{"Display all accounts registered in the blockchain"},
			Func:        doAccounts,
			Parameters: map[string]*Parameter{
				"delete": {Required: false, Description: "the account you want to delete from the blockchain"},
			},
		},
		"newaccount": {
			Description: []string{"Create a new account with a random number identified by a label"},
			Func:        doNewAccount,
			Parameters: map[string]*Parameter{
				"label": {Required: true, Description: "The label to identify the new account"},
			},
		},
		"startnode": {
			Description: []string{"Start the Node to synchronize the blockchain network with other nodes."},
			Func:        doStartNode,
			Parameters: map[string]*Parameter{
				"port": {Required: false, Description: "Set the TCP/IP port number to the listener. Default is 8085"},
			},
		},
		"startws": {
			Description: []string{"Start WebServer engine on port 8080"},
			Func:        doStartWS,
			Parameters: map[string]*Parameter{
				"port": {Required: false, Description: "Set the TCP/IP port number to the listener. Default is 8080"},
			},
		},
	}
}

func doStartWS(c *Command) {
	ws := &webserver.WebServer{}
	port := c.Parameters["port"].Value

	if len(port) > 0 {
		numPort, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}
		ws.UsePort(int(numPort))
	}

	started := make(chan bool)
	go func(w *webserver.WebServer) {
		w.Start(started)
	}(ws)

	if <-started {
		log.Println("WebServer started!")
	}
}

func doStartNode(c *Command) {
	blockchainNode := &blockchain.BlockchainNode{}
	port := c.Parameters["port"].Value

	if len(port) > 0 {
		numPort, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}
		blockchainNode.UsePort(int(numPort))
	}

	blockchainNode.StartListener()
}

func doNewAccount(c *Command) {
	account, err := blockchain.NewAccount(c.Parameters["label"].Value)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Account 0x%x created.\r\n", account.Address)
	}

	os.Exit(0)
}

func doAccounts(c *Command) {
	a := &blockchain.Account{}
	a.ListAll()
	os.Exit(0)
}

func doStartMiner(c *Command) {
	threads := c.Parameters["threads"].Value
	benchmark := c.Parameters["benchmark"].Value

	if len(threads) > 0 {
		num, err := strconv.ParseInt(threads, 10, 32)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}

		if num < 1 {
			fmt.Printf("\"%d\" is an invalid number of threads.\r\n", num)
			os.Exit(0)
		}
		NumThreads = int(num)
	}

	if len(benchmark) > 0 {
		if benchmark != "yes" && benchmark != "no" {
			fmt.Printf("\"%s\" is not a valid value. Inform \"yes\" or \"no\"\r\n", benchmark)
			os.Exit(0)
		}
		BenchmarkMode = benchmark == "yes"
	}

	miners := make([]*blockchain.Miner, NumThreads)
	bc := &blockchain.Blockchain{}
	block := bc.CurrentBlock()

	i := block.Id
	d := block.Difficulty
	h := block.Hash

	log.Printf("Current Block: [id:%d diff:%d hash:%x]\r\n", i, d, h)
	log.Printf("Started miner engine with %d threds.\r\n", len(miners))

	for i := 0; i < NumThreads; i++ {

		miners[i] = new(blockchain.Miner)
		miners[i].ThreadId = i + 1
		miners[i].StartMiner(bc, HashFoundCallBack)
	}

}

func displayHelp(c *Command) {

	makeSeparators := func(required bool) (left string, right string) {
		if required {
			return "<", ">"
		} else {
			return "[", "]"
		}
	}

	for name, item := range Commands {
		parameters := ""

		for paramName, parameter := range item.Parameters {

			leftSeparator, rightSeparator := makeSeparators(parameter.Required)
			parameters += fmt.Sprintf("%s%s:value%s ", leftSeparator, paramName, rightSeparator)
		}

		fmt.Printf("%-15sengine.exe %s %s\r\n", name, name, parameters)

		for _, description := range item.Description {
			fmt.Printf("%s%s\r\n", strings.Repeat(" ", 15), description)
		}

		if len(item.Parameters) > 0 {
			fmt.Println()

			for paramName, parameter := range item.Parameters {
				leftSeparator, rightSeparator := makeSeparators(parameter.Required)
				fmt.Printf("%s%s%s:value%s %s\r\n", strings.Repeat(" ", 15), leftSeparator, paramName, rightSeparator, parameter.Description)
			}
		}

		fmt.Println()
	}
	os.Exit(0)
}

func doSend(c *Command) {
	from := c.Parameters["from"].Value
	to := c.Parameters["to"].Value
	numAmmount, err := strconv.ParseFloat(c.Parameters["ammount"].Value, 64)
	if err != nil {
		panic(err)
	}

	if numAmmount <= 0 {
		log.Panicf("Invalid ammount: %0.f\r\n", numAmmount)
	}

	transaction := &blockchain.Transaction{}
	result, err := transaction.NewTransaction(from, to, numAmmount)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created the transaction: 0x%x\r\n", result.ID)

	os.Exit(0)
}

func doAirdrop(c *Command) {
	to := c.Parameters["to"].Value
	ammount := c.Parameters["ammount"].Value
	numAmmount, err := strconv.ParseFloat(ammount, 64)
	if err != nil {
		panic(err)
	}

	if numAmmount <= 0 || numAmmount > 10000 {
		log.Panicf("Invalid ammount: %0.f. Range allowed must be greater than 0 up to 10000.\r\n", numAmmount)
	}

	account, err := blockchain.Airdrop(to, numAmmount)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Balance: %0.8f\r\n", account.Balance)
	}
	os.Exit(0)
}

func init() {

	buildCommandList()

	if len(os.Args) <= 1 {
		return
	}

	c := Commands[os.Args[1]]
	if c == nil {
		fmt.Printf("The command \"%s\" is not recognized.\r\n", os.Args[1])
		os.Exit(0)
	}

	fnInvalidParameter := func(c *Command, paramName string) {
		fmt.Printf("Invalid argument: %s\r\n", paramName)
		displayHelp(c)
		os.Exit(0)
	}

	for i := 2; i < len(os.Args); i++ {
		items := strings.Split(os.Args[i], ":")
		if len(items) < 2 {
			fnInvalidParameter(c, os.Args[i])
		}

		parameter := c.Parameters[items[0]]
		if parameter == nil {
			fnInvalidParameter(c, items[0])
		}

		parameter.Value = items[1]
	}

	for paramName, item := range c.Parameters {
		if item.Required && len(item.Value) == 0 {
			fnInvalidParameter(c, fmt.Sprintf("parameter \"%s\" cannot be empty", paramName))
		}
	}

	c.Func(c)
}
