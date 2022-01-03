package main

import (
	"fmt"
	"github.com/meschbach/go-junk-bucket/sub"
	"os"
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
	if err := cmd.PumpToStandard(args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err.Error())
	}
}
