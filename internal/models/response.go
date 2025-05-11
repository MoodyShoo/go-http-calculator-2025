package models

import "encoding/json"

// ----- Response Interface -----
type Response interface {
	ToJSON() ([]byte, error)
}

// ----- Accepted Response -----

type AcceptedResponse struct {
	Id int64 `json:"id"`
}

func (r *AcceptedResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ----- Expressions Response -----

type ExpressionsResponse struct {
	Expressions []Expression `json:"expressions"`
}

func (r *ExpressionsResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ----- Success Response -----

type SuccessResponse struct {
	Message string `json:"message"`
}

func (r *SuccessResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
