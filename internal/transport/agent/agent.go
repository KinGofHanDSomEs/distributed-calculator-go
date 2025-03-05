package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/kingofhandsomes/distributed_calculator_go/internal/models"
)

type Agent struct {
	Port                string
	TimeAddition        time.Duration
	TimeSubtraction     time.Duration
	TimeMultiplications time.Duration
	TimeDivisions       time.Duration
	ComputingPower      int
}

func NewAgent() *Agent {
	godotenv.Load("variables.env")
	port := os.Getenv("PORT")
	intPort, err := strconv.Atoi(port)
	if port == "" || err != nil {
		port = "8080"
	}
	if intPort < 0 || intPort > 9999 {
		port = "8080"
	}
	ta, err := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	if err != nil || ta < 1 {
		ta = 1
	}
	ts, err := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	if err != nil || ts < 1 {
		ts = 1
	}
	tm, err := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	if err != nil || tm < 1 {
		tm = 1
	}
	td, err := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
	if err != nil || td < 1 {
		td = 1
	}
	cp, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil || cp < 1 {
		cp = 1
	}
	return &Agent{
		Port:                port,
		TimeAddition:        time.Duration(ta) * time.Millisecond,
		TimeSubtraction:     time.Duration(ts) * time.Millisecond,
		TimeMultiplications: time.Duration(tm) * time.Millisecond,
		TimeDivisions:       time.Duration(td) * time.Millisecond,
		ComputingPower:      cp,
	}
}

func (a *Agent) Run() {
	wg := sync.WaitGroup{}
	for {
		for i := 0; i < a.ComputingPower; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				a.TaskProcessing(i + 1)
			}()
		}
		wg.Wait()
	}
}

func (a *Agent) TaskProcessing(n int) {
	resp, err := http.Get("http://localhost:" + a.Port + "/internal/task")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}
	var res struct {
		Task *models.RespTask `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return
	}
	log.Printf("agent %d started work with task %d\n", n, res.Task.ID)
	task := res.Task
	result, duration := a.TaskCalculation(task.Arg1, task.Arg2, task.Operation)
	body, _ := json.Marshal(map[string]interface{}{
		"id":             task.ID,
		"result":         result,
		"operation_time": duration,
	})
	log.Printf("agent %d ended work with task %d, operation time: %v", n, res.Task.ID, duration)
	http.Post("http://localhost:"+a.Port+"/internal/task", "application/json", bytes.NewBuffer(body))
}

func (a *Agent) TaskCalculation(arg1, arg2 float64, oper string) (float64, time.Duration) {
	switch oper {
	case "+":
		<-time.After(a.TimeAddition)
		return arg1 + arg2, a.TimeAddition
	case "-":
		<-time.After(a.TimeSubtraction)
		return arg1 - arg2, a.TimeSubtraction
	case "*":
		<-time.After(a.TimeMultiplications)
		return arg1 * arg2, a.TimeMultiplications
	default:
		<-time.After(a.TimeDivisions)
		return arg1 / arg2, a.TimeDivisions
	}
}
