package main

import (
	"log"
	"os"
	"strconv"

	"leewill1120/yager/manager"
	"leewill1120/yager/plugin"
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
	case "plugin":
		startPlugin(os.Args[2:])
	default:
		log.Fatalf("mode(%s) doesn't support", mode)
	}
}

func startManager(args []string) {
	var (
		listenPort   int
		registerCode string
		c            chan interface{}
	)
	listenPort, _ = strconv.Atoi(args[0])
	registerCode = args[1]
	m := manager.NewManager(listenPort, registerCode)
	m.Run(c)
	log.Println("manager stopped, reason:%s\n", <-c)
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

func startPlugin(args []string) {
	StoreManagerIP := args[0]
	StoreManagerPort, _ := strconv.Atoi(args[1])
	p := plugin.NewPlugin(StoreManagerIP, StoreManagerPort)
	p.Run()
}
