package handler

import "net/http"

type MessageResponse struct {
	Message string `json:"message"`
}

func Ping(_ *http.Request) (interface{}, error) {
	return MessageResponse{Message: "pong"}, nil
}
