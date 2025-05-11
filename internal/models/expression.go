package models

import "encoding/json"

type Status string

const (
	StatusPending   Status = "pending"
	StatusComputing Status = "computing"
	StatusDone      Status = "done"
	StatusError     Status = "error"
)

type Expression struct {
	Id     int64   `json:"id"`
	Expr   string  `json:"expression"`
	Status Status  `json:"status"`
	Result float64 `json:"result"`
	Error  string  `json:"error,omitempty"`
	UserID int64   `json:"-"`
}

func (e *Expression) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
