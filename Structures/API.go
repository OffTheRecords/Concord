package Structures

type Response struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

type LoginResponse struct {
	ID     string `json:"id"`
	JwtTTL int64  `json:"jwtttl"`
}

type RegisterResponse struct {
	ID     string `json:"id"`
	JwtTTL int64  `json:"jwtttl"`
}

type RefreshResponse struct {
	ID     string `json:"id"`
	JwtTTL int64  `json:"jwtttl"`
}
