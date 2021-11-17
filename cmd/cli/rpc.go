package main

import (
	"fmt"
	"github.com/fatih/color"
	"net/rpc"
	"os"
	"sync"
)

func client(inMaintenanceMode bool) {
	var wg sync.WaitGroup

	wg.Add(1)
	c, err := rpc.Dial("tcp", "127.0.0.1:"+os.Getenv("RPC_PORT"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connected...")
	var result string
	err = c.Call("RPCServer.MaintenanceMode", inMaintenanceMode, &result)
	if err != nil {
		exitGracefully(err)
	} else {
		color.Yellow(result)
	}
	wg.Done()
}
