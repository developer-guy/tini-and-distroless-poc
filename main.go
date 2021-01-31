package main

import (
	"log"
	"time"

	"github.com/mitchellh/go-ps"
)

func main() {
	list, err := ps.Processes()
	if err != nil {
		panic(err)
	}
	for _, p := range list {
		log.Printf("Process %s with PID %d and PPID %d", p.Executable(), p.Pid(), p.PPid())
	}

	time.Sleep(3600 * time.Second)
}
