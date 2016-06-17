package main

import (
	"log"
	"os"
	"strconv"

	"leewill1120/yager/client"
	"leewill1120/yager/slave"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("no enough arguments.")
	}
	mode := os.Args[1]

	switch mode {
	case "master":

	case "slave":
		startSlave()
	default:
		startClient(os.Args[1:])
	}
}

func startSlave() {
	s := slave.NewSlave()
	if s == nil {
		log.Fatal("failed to create slave, exit.")
	} else {
		s.Run()
	}
}

func startClient(args []string) {
	cmd := args[0]
	c := client.NewClient(args[1], args[2])
	switch cmd {
	case "getBlock":
		size, _ := strconv.ParseFloat(args[3], 64)
		c.CmdGetBlock(size)
	case "removeBlock":
		devPath := args[3]
		c.CmdRemoveBlock(devPath)
	default:
		log.Fatal("command " + cmd + " not found.")
	}
}
