package main 

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const PORT = ":42069"

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
			readChan := getLinesChannel(conn)
			for msgs := range readChan {
				fmt.Printf("read: %s\n", msgs)
			}
		} ()
			
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lineChan := make(chan string)

	go func ()  {
		defer f.Close()
		defer fmt.Printf("Closing connection\n")
		defer close(lineChan)

		var line string = "";
		for {
			word := make([]byte, 8)
			_, err := f.Read(word)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error Reading file")
				close(lineChan)
				return
			}

			strSlice := strings.Split(string(word), "\n")
			line += strSlice[0]

			for index := 1; index < len(strSlice); index++ {
				lineChan <- line
				line = strSlice[index]
			}
		}
		lineChan <- line
	}()
	
	return lineChan
}