package models

type Request struct {
	Expression string `json:"expression"`
}

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
