package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 8080

func main() {
	// ser, err := server.Serve(port, handlerFunc)
	// if err != nil {
	// 	log.Fatalf("Error starting server: %v", err)
	// }
	// defer ser.Close()
	// log.Println("Server started on port", port)

	ser1, err := server.Serve(port, videoHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer ser1.Close()

	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handlerFunc(writer *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		resp, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")))
		if err != nil {
			writer.WriteResponse(400, "Failed to fetch from httpbin")
			return
		}

		header := headers.NewHeaders()
		for key, values := range resp.Header {
			for _, value := range values {
				header.Put(key, value)
			}
		}
		header.Remove("content-length")
		header.Put("Transfer-Encoding", "chunked")
		header.Put("Trailer", "X-Content-SHA256")
		header.Put("Trailer", "X-Content-Length")

		writer.WriteStatusLine(response.StatusCode(resp.StatusCode))
		writer.WriteHeaders(header)

		var body []byte
		for {
			buffer := make([]byte, 128)
			n, err := resp.Body.Read(buffer)
			body = append(body, buffer[:n]...)
			if err == io.EOF {
				writer.WriteChunkedBodyDone()
				break
			} else if err == nil {
				writer.WriteChunkedBody(buffer[:n])
			}
		}
		hash := sha256.Sum256(body)

		trailer := headers.NewHeaders()
		trailer.Put("X-Content-SHA256", fmt.Sprintf("%x", hash))
		trailer.Put("X-Content-Length", fmt.Sprintf("%d",len(body)))
		writer.WriteTrailers(trailer)
		return
	}

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
		header.Replace("content-type", "text/html")
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
		header.Replace("content-type", "text/html")
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
		header.Replace("content-type", "text/html")
		writer.WriteHeaders(header)
		writer.WriteBody(responseBody)
	}
}

func videoHandler(writer *response.Writer, req *request.Request) {
	if req.RequestLine.Method == "GET" && req.RequestLine.RequestTarget == "/video" {
		video, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			fmt.Printf("Failed to read video. %s", err.Error())
		}

		writer.WriteStatusLine(200)

		header := response.GetDefaultHeader(0)
		header.Remove("content-length")
		header.Replace("content-type", "video/mp4")
		header.Put("Transfer-Encoding", "chunked")
		header.Put("Trailer", "X-Content-SHA256")
		header.Put("Trailer", "X-Content-Length")
		writer.WriteHeaders(header)

		data := bytes.NewBuffer(video)
		for {
			buffer := make([]byte, 128)
			n, err := data.Read(buffer)
			if err != nil {
				break
			}
			writer.WriteChunkedBody(buffer[:n])
		}
		writer.WriteChunkedBodyDone()
		hash := sha256.Sum256(video)

		trailer := headers.NewHeaders()
		trailer.Put("X-Content-SHA256", fmt.Sprintf("%x", hash))
		trailer.Put("X-Content-Length", fmt.Sprintf("%d",len(video)))
		writer.WriteTrailers(trailer)
		
	}
}