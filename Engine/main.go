package main

import (
	"engine/blockchain"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	Version       = "1.0.0"
	BuildTime     = time.Now().Format(time.RFC3339)
	NumThreads    = runtime.NumCPU()
	mutexTicker   = new(sync.Mutex)
	tickerCounter = 0
	BenchmarkMode = false
)

func tick() {
	for {
		h := tickerCounter / 3600
		m := tickerCounter / 60 % 60
		s := tickerCounter % 60
		fmt.Printf("\r%d:%02d:%02d", h, m, s)
		time.Sleep(time.Second)
		tickerCounter++
	}
}

func HashFoundCallBack(hash *blockchain.HashBlock, nonce *blockchain.Nonce) {
	mutexTicker.Lock()
	defer mutexTicker.Unlock()
	tickerCounter = 0
}

func main() {
	fmt.Printf("HSN Blockchain 2.0 Engine %s %s\r\n", Version, BuildTime)

	buildCommandList()

	if len(os.Args) == 1 {
		displayHelp(nil)
		os.Exit(0)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go tick()
	wg.Wait()
}
