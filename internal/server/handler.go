package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message string
}

type Handler func(*response.Writer, *request.Request)

func WriteError(w io.Writer, handlerError *HandlerError) error {
	_, err := fmt.Fprintf(w, "Server responded with status code %d and response message %s", handlerError.StatusCode, handlerError.Message)
	return err
}