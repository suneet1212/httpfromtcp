package main 

import (
	"fmt"
	"log"
	"net"
	"httpfromtcp/internal/request"
)

const PORT = ":8080"

func main() {
	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatal("Could not connect to the address")
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Can't accept any more connections on %s", PORT)
			break
		}

		fmt.Printf("Connected on port %s\n", PORT)
		
		go func ()  {
			req, err := request.RequestFromReader(conn)
			if err != nil {
				log.Printf("Unable to parse from connection")
				return
			}

			fmt.Printf("Request line:\n - Method: %s\n - Target: %s\n - Version: %s\n",
				req.RequestLine.Method,
				req.RequestLine.RequestTarget,
				req.RequestLine.HttpVersion,
			)
			fmt.Print("Headers:\n")
			for k, v := range req.Headers {
				fmt.Printf(" - %s: %s\n", k, v)
			}
			fmt.Printf("Body:\n%s\n", req.Body)

		} ()
			
	}
}