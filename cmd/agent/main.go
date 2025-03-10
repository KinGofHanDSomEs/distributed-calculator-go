package main

import (
	"sync"

	"github.com/kingofhandsomes/distributed_calculator_go/internal/transport/agent"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.NewAgent().Run()
	}()
	wg.Wait()
}
