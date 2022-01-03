package main

import (
	"fmt"
	"github.com/meschbach/go-junk-bucket/sub"
	"os"
	"sync"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		if _, err := fmt.Fprintf(os.Stderr, "Usage requires the following format: <program> [<args>]+\n"); err != nil {
			panic(err)
		}
		return
	}
	cmd := sub.NewSubcommand(args[1], args[2:])

	var pumpsDone sync.WaitGroup

	toClose := make([]chan string, 0)
	buildPump := func(name string) chan string {
		pipe := make(chan string, 8)
		toClose = append(toClose, pipe)
		pumpsDone.Add(1)
		go sub.PumpPrefixedChannel(name, pipe, &pumpsDone)
		return pipe
	}

	stdout := buildPump("<<stdout>>")
	stderr := buildPump("<<stderr>>")

	if err := cmd.Run(stdout, stderr); err != nil {
		panic(err)
	}
	for _, pipe := range toClose {
		close(pipe)
	}
	pumpsDone.Wait()
}
