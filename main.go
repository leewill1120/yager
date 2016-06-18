package main

import (
	"log"
	"os"
	"strconv"

	"leewill1120/yager/client"
	"leewill1120/yager/manager"
	"leewill1120/yager/worker"
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
	case "manager":
		startManager(os.Args[2:])
	case "slave":
		startWorker("slave", os.Args[2:])
	case "standalone":
		startWorker("standalone", os.Args[2:])
	default:
		startClient(os.Args[2:])
	}
}

func startManager(args []string) {
	var (
		listenPort   int
		registerCode string
	)
	listenPort, _ = strconv.Atoi(args[0])
	registerCode = args[1]
	m := manager.NewManager(listenPort, registerCode)
	m.Run()
}
func startWorker(mode string, args []string) {
	var (
		masterIP     string
		registerCode string
		masterPort   int
		listenPort   int
	)

	switch mode {
	case "slave":
		masterIP = args[0]
		masterPort, _ = strconv.Atoi(args[1])
		listenPort, _ = strconv.Atoi(args[2])
		registerCode = args[3]
	case "standalone":
		listenPort, _ = strconv.Atoi(args[0])
	}

	s := worker.NewWorker(mode, masterIP, masterPort, listenPort, registerCode)
	if s == nil {
		log.Fatalf("failed to create %s worker, exit.", mode)
	} else {
		c := make(chan error)
		s.Run(c)
		log.Printf("server stopped, reason: %s\n", <-c)
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
