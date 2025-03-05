package main

import (
	"sync"

	"github.com/kingofhandsomes/distributed_calculator_go/internal/transport/agent"
	"github.com/kingofhandsomes/distributed_calculator_go/internal/transport/orchestrator"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		orchestrator.NewOrchestrator().Run()
	}()
	go func() {
		defer wg.Done()
		agent.NewAgent().Run()
	}()
	wg.Wait()
}
