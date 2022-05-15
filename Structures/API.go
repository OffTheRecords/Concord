package Structures

type Response struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

type LoginResponse struct {
	ID string `json:"id"`
}

type RegisterResponse struct {
	ID string `json:"id"`
}
