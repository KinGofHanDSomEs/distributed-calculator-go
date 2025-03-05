package orchestrator_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/kingofhandsomes/distributed_calculator_go/internal/models"
	"github.com/kingofhandsomes/distributed_calculator_go/internal/transport/orchestrator"
)

func TestAddExpression(t *testing.T) {
	t.Parallel()

	log.SetOutput(io.Discard)
	defer log.SetOutput(log.Writer())

	o := orchestrator.NewOrchestrator()

	testCases := []struct {
		name               string
		expr               string
		expectedStatusCode int
	}{
		{
			name:               "simple expression",
			expr:               "2+2+2",
			expectedStatusCode: 201,
		},
		{
			name:               "all operations",
			expr:               "1+(2-3)*4/5",
			expectedStatusCode: 201,
		},
		{
			name:               "all operations with real numbers",
			expr:               "1+(2.2-3)*4.45/5",
			expectedStatusCode: 201,
		},
		{
			name:               "all operations with negative real numbers",
			expr:               "1+(-2.2)-3*4.45/(-5)",
			expectedStatusCode: 201,
		},
		{
			name:               "division by zero",
			expr:               "2+6-7/0",
			expectedStatusCode: 422,
		},
		{
			name:               "invalid character found",
			expr:               "2+2+$2",
			expectedStatusCode: 422,
		},
		{
			name:               "real number is written incorrectly",
			expr:               "2.1+4*7.5.2",
			expectedStatusCode: 422,
		},
		{
			name:               "extra opening bracket",
			expr:               "(2+3)*(5+3",
			expectedStatusCode: 422,
		},
		{
			name:               "extra closing bracket",
			expr:               "2+3*(5+6-1))",
			expectedStatusCode: 422,
		},
	}
	for _, ts := range testCases {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			t.Parallel()

			reqBody, _ := json.Marshal(models.ReqAddExpr{Expression: ts.expr})
			r := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			o.AddExpression(w, r)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != ts.expectedStatusCode {
				t.Errorf("invalid status code: got %v want %v", res.StatusCode, ts.expectedStatusCode)
			}
		})
	}
}

func TestGetExpressions(t *testing.T) {
	t.Parallel()

	log.SetOutput(io.Discard)
	defer log.SetOutput(log.Writer())

	o := orchestrator.NewOrchestrator()

	o.Exprs[1] = &orchestrator.Expression{
		ID:        1,
		Status:    "resolved",
		Result:    4,
		Body:      "2+2",
		EndTaskID: 1,
	}
	o.Exprs[2] = &orchestrator.Expression{
		ID:        2,
		Status:    "not resolved",
		Body:      "1+2+3",
		EndTaskID: 3,
	}

	r := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	w := httptest.NewRecorder()

	o.GetExpressions(w, r)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("invalid status code: got %v want %v", res.StatusCode, http.StatusOK)
	}
}

func TestGetExpressionByID(t *testing.T) {
	t.Parallel()

	log.SetOutput(io.Discard)
	defer log.SetOutput(log.Writer())

	o := orchestrator.NewOrchestrator()
	o.Exprs[1] = &orchestrator.Expression{ID: 1, Status: "not resolved", Body: "2 + 2"}

	testCases := []struct {
		name               string
		id                 string
		expectedStatusCode int
	}{
		{"correct id", "1", 200},
		{"incorrect id", "2", 404},
	}
	for _, ts := range testCases {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest("GET", "/api/v1/expressions/1", nil)
			r = mux.SetURLVars(r, map[string]string{"id": ts.id})
			w := httptest.NewRecorder()

			o.GetExpressionByID(w, r)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != ts.expectedStatusCode {
				t.Errorf("invalid status code: got %v want %v", res.StatusCode, ts.expectedStatusCode)
			}
		})
	}
}

func TestTaskHandler(t *testing.T) {
	t.Parallel()

	log.SetOutput(io.Discard)
	defer log.SetOutput(log.Writer())

	o := orchestrator.NewOrchestrator()
	o.Tasks[1] = &orchestrator.Task{ID: 1, Arg1: 2, Arg2: 2, Operation: "+", Status: "untouched"}

	testCasesGet := []struct {
		name               string
		expectedID         int
		expectedStatusCode int
	}{
		{
			name:               "get task status - OK",
			expectedStatusCode: 200,
		},
		{
			name:               "get task status - not found",
			expectedStatusCode: 404,
		},
	}
	for _, ts := range testCasesGet {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodGet, "/internal/task", nil)
			w := httptest.NewRecorder()

			o.TaskHandler(w, r)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != ts.expectedStatusCode {
				t.Fatalf("invalid status code: got %v want %v", res.StatusCode, ts.expectedStatusCode)
			}
		})
	}

	testCasesPost := []struct {
		name               string
		id                 int
		result             float64
		operationTime      time.Duration
		expectedStatusCode int
	}{
		{
			name:               "post task status - OK",
			id:                 1,
			result:             4,
			operationTime:      time.Millisecond,
			expectedStatusCode: 200,
		},
		{
			name:               "post task status - not found",
			id:                 2,
			result:             2,
			operationTime:      time.Millisecond,
			expectedStatusCode: 404,
		},
	}
	for _, ts := range testCasesPost {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			t.Parallel()
			jsonBytes, _ := json.Marshal(models.ReqTask{ID: ts.id, Result: ts.result, OperationTime: ts.operationTime})

			r := httptest.NewRequest(http.MethodPost, "/intarnal/task", bytes.NewBuffer(jsonBytes))
			w := httptest.NewRecorder()

			o.TaskHandler(w, r)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != ts.expectedStatusCode {
				t.Fatalf("invalid status code: got %v want %v", res.StatusCode, ts.expectedStatusCode)
			}
		})
	}
	t.Run("post task status - invalid data", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/internal/task", nil)
		w := httptest.NewRecorder()

		o.TaskHandler(w, r)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("invalid status code: got %v want %v", res.StatusCode, http.StatusUnprocessableEntity)
		}
	})
}
