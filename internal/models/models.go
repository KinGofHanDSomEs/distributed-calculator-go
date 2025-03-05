package models

import (
	"time"
)

type ReqAddExpr struct {
	Expression string `json:"expression"`
}

type RespAddExpr struct {
	ID int `json:"id"`
}

type RespExpr struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type ReqTask struct {
	ID            int           `json:"id"`
	Result        float64       `json:"result"`
	OperationTime time.Duration `json:"operation_time"`
}

type RespTask struct {
	ID            int           `json:"id"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}
