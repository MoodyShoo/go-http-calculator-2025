package models

import "encoding/json"

type Task struct {
	Id            int     `json:"id"`
	ExpressionId  int     `json:"expression_id"`
	Arg1          string  `json:"arg1"`
	Arg2          string  `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
	Status        Status  `json:"status"`
	Result        float64 `json:"result,omitempty"`
	Error         string  `json:"error,omitempty"`
}

func (r *Task) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

type TaskResult struct {
	Id     int     `json:"id"`
	Result float64 `json:"result"`
	Error  string  `json:"error,omitempty"`
}

func (r *TaskResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
