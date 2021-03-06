package sub

import (
	"fmt"
	"os/exec"
	"sync"
)

type Option interface {
	Customize(cmd *exec.Cmd)
}

type Subcommand struct {
	programName      string
	programArguments []string
	Options []Option
}

func NewSubcommand(programName string, args []string) *Subcommand {
	return &Subcommand{
		programName:      programName,
		programArguments: args,
	}
}

func (s *Subcommand) WithOption(opt Option)  {
	s.Options = append(s.Options, opt)
}

func (s *Subcommand) Interact(stdin <-chan string, stdout chan<- string, stderr chan<- string ) error {
	var completedReading sync.WaitGroup
	var readyGate sync.WaitGroup

	cmd := exec.Command(s.programName, s.programArguments...)
	for _, option := range s.Options {
		option.Customize(cmd)
	}

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	readyGate.Add(1)
	completedReading.Add(1)
	go func() {
		readyGate.Done()
		for in := range stdin {
			fmt.Printf("<<stdin>> %q\n", in)
			withNewLine := in + "\n"
			asBytes := []byte(withNewLine)
			_, err := stdinPipe.Write(asBytes)
			if err != nil { panic(err) }
		}
		err := stdinPipe.Close()
		if err != nil { panic(err) }
		completedReading.Done()
	}()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil { return err }
	stderrPipe, err := cmd.StderrPipe()
	if err != nil { return err }

	completedReading.Add(1)
	readyGate.Add(1)
	go PumpLines(&GivePumpSource{Source: stdoutPipe}, stdout, &completedReading, &readyGate)

	completedReading.Add(1)
	readyGate.Add(1)
	go PumpLines(&GivePumpSource{Source: stderrPipe}, stderr, &completedReading, &readyGate)

	readyGate.Wait()
	err = cmd.Run()
	stdoutPipe.Close()
	stderrPipe.Close()

	completedReading.Wait()
	return err
}

func (s *Subcommand) Run(stdout chan<- string, stderr chan<- string) error {
	stdin := make(chan string)
	close(stdin)

	return s.Interact(stdin,stdout,stderr)
}
