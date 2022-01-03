package sub

import (
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

func (s *Subcommand) Run(stdout chan<- string, stderr chan<- string) error {
	cmd := exec.Command(s.programName, s.programArguments...)
	for _, option := range s.Options {
		option.Customize(cmd)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := stdin.Close(); err != nil {
		return err
	}

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	var completedReading sync.WaitGroup
	var readyGate sync.WaitGroup

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
