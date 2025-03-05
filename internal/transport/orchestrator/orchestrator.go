package orchestrator

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kingofhandsomes/distributed_calculator_go/internal/errors"
	"github.com/kingofhandsomes/distributed_calculator_go/internal/models"
)

type Orchestrator struct {
	Port         string
	Exprs        map[int]*Expression
	Tasks        map[int]*Task
	Mu           sync.Mutex
	IdExpr       int
	IdTask       int
	IdTaskSolved int
}

func NewOrchestrator() *Orchestrator {
	godotenv.Load("variables.env")
	port := os.Getenv("PORT")
	intPort, err := strconv.Atoi(port)
	if port == "" || err != nil {
		port = "8080"
	}
	if intPort < 0 || intPort > 9999 {
		port = "8080"
	}
	return &Orchestrator{
		Port:   port,
		Exprs:  make(map[int]*Expression),
		Tasks:  make(map[int]*Task),
		IdExpr: 1,
		IdTask: 1,
	}
}

type Expression struct {
	ID        int
	Status    string
	Result    float64
	Body      string
	EndTaskID int
}

type Task struct {
	ID            int
	Arg1          float64
	Arg2          float64
	Operation     string
	OperationTime time.Duration
	Status        string
	Result        float64
}

var (
	precedence = map[rune]int{
		'+': 1,
		'-': 1,
		'*': 2,
		'/': 2,
	}
)

func ToPolishNotation(expression string) ([]string, error) {
	output := []string{}
	stack := []rune{}
	i := 0
	for i < len(expression) {
		char := rune(expression[i])
		if unicode.IsDigit(char) || (char == '-' && (i == 0 || expression[i-1] == '(')) {
			start := i
			if char == '-' {
				i++
			}
			for i < len(expression) && (unicode.IsDigit(rune(expression[i])) || expression[i] == '.') {
				i++
			}
			number := expression[start:i]
			output = append(output, number)
			continue
		}
		if !(char == '+' || char == '-' || char == '*' || char == '/' || char == '(' || char == ')') {
			return nil, errors.ErrInvalidSymbol
		}
		if char == '+' || char == '-' || char == '*' || char == '/' {
			for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[char] {
				output = append(output, string(stack[len(stack)-1]))
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, char)
		} else if char == '(' {
			stack = append(stack, char)
		} else if char == ')' {
			for len(stack) > 0 && stack[len(stack)-1] != '(' {
				output = append(output, string(stack[len(stack)-1]))
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, errors.ErrClosingBracket
			}
			stack = stack[:len(stack)-1]
		}
		i++
	}
	for len(stack) > 0 {
		if stack[len(stack)-1] == '(' {
			return nil, errors.ErrOpeningBracket
		}
		output = append(output, string(stack[len(stack)-1]))
		stack = stack[:len(stack)-1]
	}
	return output, nil
}

func (o *Orchestrator) AddExpression(w http.ResponseWriter, r *http.Request) {
	var req models.ReqAddExpr
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Expression == "" {
		log.Println("incorrect processing expression entered")
		http.Error(w, errors.ErrInvalidData.Error(), http.StatusUnprocessableEntity)
		return
	}

	expr := strings.ReplaceAll(req.Expression, " ", "")

	o.Mu.Lock()
	defer o.Mu.Unlock()

	rpn, err := ToPolishNotation(expr)
	if err != nil {
		log.Printf("it is impossible to create a reverse polish notation for the expression: %s\n", expr)
		http.Error(w, errors.ErrInvalidData.Error(), http.StatusUnprocessableEntity)
		return
	}
	stack, tasks, i := []float64{}, []*Task{}, -1

	for _, oper := range rpn {
		num, err := strconv.ParseFloat(oper, 64)
		if err != nil {
			if len(stack) < 2 {
				log.Printf("it is impossible to create a reverse polish notation for the expression: %s\n", expr)
				http.Error(w, errors.ErrInvalidData.Error(), http.StatusUnprocessableEntity)
				return
			}
			if !(oper == "+" || oper == "-" || oper == "*" || oper == "/") {
				log.Printf("it is impossible to create a reverse polish notation for the expression: %s\n", expr)
				http.Error(w, errors.ErrInvalidData.Error(), http.StatusUnprocessableEntity)
				return
			}
			i++
			arg1, arg2 := stack[len(stack)-2], stack[len(stack)-1]
			stack = stack[:len(stack)-2]
			switch oper {
			case "+":
				stack = append(stack, arg1+arg2)
			case "-":
				stack = append(stack, arg1-arg2)
			case "*":
				stack = append(stack, arg1*arg2)
			case "/":
				if arg2 == 0 {
					log.Printf("division by zero error for the expression: %s\n", expr)
					http.Error(w, errors.ErrInvalidData.Error(), http.StatusUnprocessableEntity)
					return
				}
				stack = append(stack, arg1/arg2)
			}

			tasks = append(tasks, &Task{
				ID:        o.IdTask + i,
				Arg1:      arg1,
				Arg2:      arg2,
				Operation: oper,
				Status:    "untouched",
				Result:    0,
			})
		} else {
			stack = append(stack, num)
		}
	}
	if len(stack) != 1 {
		log.Printf("it is impossible to create a reverse polish notation for the expression: %s\n", expr)
		http.Error(w, errors.ErrInvalidData.Error(), http.StatusUnprocessableEntity)
		return
	}

	expression := &Expression{
		ID:        o.IdExpr,
		Status:    "not resolved",
		Body:      expr,
		EndTaskID: o.IdTask + i,
	}
	o.Exprs[expression.ID] = expression
	o.IdExpr++
	o.IdTask += i + 1

	for _, task := range tasks {
		o.Tasks[task.ID] = task
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(models.RespAddExpr{ID: expression.ID}); err != nil {
		log.Println("server returned an error")
		http.Error(w, errors.ErrServerSide.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("successful addition of an expression - %s, with id - %d\n", expression.Body, expression.ID)
}

func (o *Orchestrator) GetExpressions(w http.ResponseWriter, r *http.Request) {
	o.Mu.Lock()
	defer o.Mu.Unlock()

	var resp []models.RespExpr
	for _, expr := range o.Exprs {
		resp = append(resp, models.RespExpr{
			ID:     expr.ID,
			Status: expr.Status,
			Result: expr.Result,
		})
	}
	if err := json.NewEncoder(w).Encode(map[string][]models.RespExpr{"expressions": resp}); err != nil {
		log.Println("server returned an error")
		http.Error(w, errors.ErrServerSide.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("expressions were successfully output")
}

func (o *Orchestrator) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("an incorrect id - %s was requested for the expression\n", idStr)
		http.Error(w, errors.ErrNotFound.Error(), http.StatusNotFound)
		return
	}

	o.Mu.Lock()
	defer o.Mu.Unlock()

	expr, ok := o.Exprs[id]
	if !ok {
		log.Printf("an expression with an invalid id - %d was requested\n", id)
		http.Error(w, errors.ErrNotFound.Error(), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]models.RespExpr{"expression": models.RespExpr{
		ID:     expr.ID,
		Status: expr.Status,
		Result: expr.Result,
	}}); err != nil {
		log.Println("server returned an error")
		http.Error(w, errors.ErrServerSide.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("expression: %s, with id: %d, was successfully output\n", expr.Body, expr.ID)
}

func (o *Orchestrator) TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		o.Mu.Lock()
		defer o.Mu.Unlock()
		o.IdTaskSolved++
		task, ok := o.Tasks[o.IdTaskSolved]
		if !ok {
			o.IdTaskSolved--
			http.Error(w, errors.ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]models.RespTask{"task": {ID: task.ID, Arg1: task.Arg1, Arg2: task.Arg2, Operation: task.Operation, OperationTime: task.OperationTime}}); err != nil {
			log.Println("server returned an error")
			http.Error(w, errors.ErrServerSide.Error(), http.StatusInternalServerError)
			return
		}
		task.Status = "solved"
	case http.MethodPost:
		var req models.ReqTask
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("an incorrect issue result structure was sent.")
			http.Error(w, errors.ErrInvalidData.Error(), http.StatusUnprocessableEntity)
			return
		}
		o.Mu.Lock()
		defer o.Mu.Unlock()

		task, ok := o.Tasks[req.ID]
		if !ok {
			log.Printf("the problem with the id - %d was not found to be solved\n", req.ID)
			http.Error(w, errors.ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		task.Result = req.Result
		task.Status = "resolved"
		task.OperationTime = req.OperationTime
		o.Tasks[task.ID] = task
		for _, expr := range o.Exprs {
			if expr.EndTaskID == task.ID {
				log.Printf("expression %d was successfully calculated\n", expr.ID)
				expr.Status = "resolved"
				expr.Result = task.Result
				return
			}
		}
	}
}

func (o *Orchestrator) Run() {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/calculate", o.AddExpression).Methods("POST")
	r.HandleFunc("/api/v1/expressions", o.GetExpressions).Methods("GET")
	r.HandleFunc("/api/v1/expressions/{id}", o.GetExpressionByID).Methods("GET")
	r.HandleFunc("/internal/task", o.TaskHandler).Methods("GET", "POST")
	log.Printf("the server is running on the port: %s\n", o.Port)
	http.ListenAndServe(":"+o.Port, r)
}
