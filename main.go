package main

import (
	"log"
	"os"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"
)

func main() {
	//cmd := exec.Command("sleep", "10")
	//cmdOutput := &bytes.Buffer{}
	//cmd.Stdout = cmdOutput
	//
	//err := cmd.Start()
	//if err != nil {
	//	// Run could also return this error and push the program
	//	// termination decision to the `main` method.
	//	log.Fatal(err)
	//}
	//
	//result := make(chan error, 1)
	//go func() {
	//	result <- cmd.Wait()
	//}()

	//cmd, err := exec.LookPath("sleep")
	//if err != nil {
	//	log.Fatal(err)
	//}

	// You need to run the program as  root to do this

	var cred = &syscall.Credential{Uid: 0, Gid: 0, Groups: []uint32{}, NoSetGroups: false}

	// the Noctty flag is used to detach the process from parent tty
	var sysproc = &syscall.SysProcAttr{Credential: cred}
	var attr = os.ProcAttr{
		Dir: ".",
		Env: os.Environ(),
		Sys: sysproc,
	}
	result := make(chan bool, 1)

	go func() {
		process, err := os.StartProcess("/bin/sleep", []string{"/bin/sleep", "15"}, &attr)
		if err == nil {
			state, err := process.Wait()
			if err != nil {
				panic(err)
			}
			result <- state.Success()
		} else {
			panic(err)
		}
	}()

	list, err := ps.Processes()

	if err != nil {
		panic(err)
	}

	for _, p := range list {
		log.Printf("Process %s with PID %d and PPID %d", p.Executable(), p.Pid(), p.PPid())
	}

	select {
	case r := <-result:
		if !r {
			log.Fatal("could not execute sleep command,error:", err)
		}
		log.Println("everything worked, closing channel")
		close(result)
	case <-time.After(20 * time.Second):
		log.Println("timeout")
	}
}
