package sub

import (
	"fmt"
	"sync"
)

func PumpPrefixedChannel(prefix string, source <-chan string, onDone *sync.WaitGroup)  {
	defer onDone.Done()
	for {
		line, ok := <- source
		if !ok {
			return
		}
		fmt.Printf("%s %s\n", prefix, line)
	}
}
