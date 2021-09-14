package sub

import (
	"bufio"
	"io"
	"sync"
)

type PumpSource interface {
	ObtainSource() (io.ReadCloser, error)
}

type GivePumpSource struct {
	Source io.ReadCloser
}

func (g *GivePumpSource) ObtainSource() (io.ReadCloser, error) {
	return g.Source, nil
}

func PumpLines(inlet PumpSource, sink chan<- string, onDone *sync.WaitGroup, readyGate *sync.WaitGroup) {
	source, err := inlet.ObtainSource()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(source)
	scanner.Split(bufio.ScanLines)
	readyGate.Done()
	for scanner.Scan() {
		text := scanner.Text()
		sink <- text
	}
	onDone.Done()
}
