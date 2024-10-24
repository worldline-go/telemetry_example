package model

type Message struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Service struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}
