package main

import (
	"sync"

	"github.com/kingofhandsomes/distributed_calculator_go/internal/transport/orchestrator"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		orchestrator.NewOrchestrator().Run()
	}()
	wg.Wait()
}
