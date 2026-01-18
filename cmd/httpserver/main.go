package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	ser, err := server.Serve(port, handlerFunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer ser.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handlerFunc(writer *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		writer.WriteStatusLine(400)

		responseBody := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
		header := response.GetDefaultHeader(len(responseBody))
		header.PutOverWrite("content-type", "text/html")
		writer.WriteHeaders(header)
		writer.WriteBody(responseBody)

	case "/myproblem":
		writer.WriteStatusLine(500)

		responseBody := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
		header := response.GetDefaultHeader(len(responseBody))
		header.PutOverWrite("content-type", "text/html")
		writer.WriteHeaders(header)
		writer.WriteBody(responseBody)

	default:
		writer.WriteStatusLine(200)

		responseBody := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
		header := response.GetDefaultHeader(len(responseBody))
		header.PutOverWrite("content-type", "text/html")
		writer.WriteHeaders(header)
		writer.WriteBody(responseBody)
	}
}