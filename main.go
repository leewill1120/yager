package main

import (
	"log"
	"os"

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
	case "client":

	default:
		log.Fatalf("mode(%s) not supported", mode)
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
