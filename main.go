package main

import (
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"leewill1120/yager/manager"
	"leewill1120/yager/plugin"
	"leewill1120/yager/worker"
)

func init() {
	log.SetLevel(log.DebugLevel)
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
	log.WithFields(log.Fields{
		"reason": <-c,
	}).Info("manager stopped.")
}
func startWorker(mode string, args []string) {
	var (
		masterIP     string
		registerCode string
		masterPort   int
		listenPort   int
		protoList    map[string]bool
	)

	switch mode {
	case "slave":
		masterIP = args[0]
		masterPort, _ = strconv.Atoi(args[1])
		listenPort, _ = strconv.Atoi(args[2])
		registerCode = args[3]
		for _, arg := range args[4:] {
			protoList[arg] = true
		}
	case "standalone":
		listenPort, _ = strconv.Atoi(args[0])
		for _, arg := range args[1:] {
			protoList[arg] = true
		}
	}

	if 0 == len(protoList) {
		log.Fatal("protocol list is empty, exit.")
	}

	w := worker.NewWorker(mode, masterIP, masterPort, listenPort, registerCode, protoList)
	if w == nil {
		log.WithFields(log.Fields{
			"mode": mode,
		}).Fatal("failed to create worker, exit.")
	} else {
		c := make(chan error)
		w.Run(c)
		log.WithFields(log.Fields{
			"reason": <-c,
		}).Info("worker stopped.")
	}
}

func startPlugin(args []string) {
	StoreManagerIP := args[0]
	StoreManagerPort, _ := strconv.Atoi(args[1])
	p := plugin.NewPlugin(StoreManagerIP, StoreManagerPort)
	p.Run()
}
