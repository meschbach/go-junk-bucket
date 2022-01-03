package sub

import (
	"fmt"
	"sync"
)

//PumpToStandard will run the given command
func (s *Subcommand) PumpToStandard(title string) error {
	var pumpsDone sync.WaitGroup

	toClose := make([]chan string, 0)
	buildPump := func(name string) chan string {
		pipe := make(chan string, 8)
		toClose = append(toClose, pipe)
		pumpsDone.Add(1)
		go PumpPrefixedChannel(name, pipe, &pumpsDone)
		return pipe
	}

	cleanUp := func() {
		for _, pipe := range toClose {
			close(pipe)
		}
	}

	stdoutName := fmt.Sprintf("<<%s.stdout>>", title)
	stdout := buildPump(stdoutName)

	stderrName := fmt.Sprintf("<<%s.stderr>>", title)
	stderr := buildPump(stderrName)

	if err := s.Run(stdout, stderr); err != nil {
		cleanUp()
		return err
	}
	cleanUp()
	pumpsDone.Wait()
	return nil
}
