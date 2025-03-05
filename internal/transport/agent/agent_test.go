package agent_test

import (
	"io"
	"log"
	"testing"
	"time"

	"github.com/kingofhandsomes/distributed_calculator_go/internal/transport/agent"
)

func TestTaskCalculation(t *testing.T) {
	t.Parallel()

	log.SetOutput(io.Discard)
	defer log.SetOutput(log.Writer())

	a := agent.NewAgent()

	testCases := []struct {
		name                  string
		arg1                  float64
		arg2                  float64
		operation             string
		expectedOperationTime time.Duration
		expectedResult        float64
	}{
		{
			name:                  "addition",
			arg1:                  2,
			arg2:                  2,
			operation:             "+",
			expectedOperationTime: a.TimeAddition,
			expectedResult:        4,
		},
		{
			name:                  "subtraction",
			arg1:                  2,
			arg2:                  2,
			operation:             "-",
			expectedOperationTime: a.TimeSubtraction,
			expectedResult:        0,
		},
		{
			name:                  "multiplications",
			arg1:                  2,
			arg2:                  2,
			operation:             "*",
			expectedOperationTime: a.TimeMultiplications,
			expectedResult:        4,
		},
		{
			name:                  "divisions",
			arg1:                  2,
			arg2:                  2,
			operation:             "/",
			expectedOperationTime: a.TimeDivisions,
			expectedResult:        1,
		},
	}

	for _, ts := range testCases {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			t.Parallel()
			res, timeOper := a.TaskCalculation(ts.arg1, ts.arg2, ts.operation)
			log.Println(timeOper)
			if res != ts.expectedResult {
				t.Fatalf("invalid result: got %v want %v\n", res, ts.expectedResult)
			}

			if timeOper != ts.expectedOperationTime {
				t.Fatalf("invalid operation time: got %v want %v\n", timeOper, ts.expectedOperationTime)
			}
		})
	}
}
