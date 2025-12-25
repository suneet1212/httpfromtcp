package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

const portNumber string = "42069"

func main() {
    listner, err := net.Listen("tcp", ":"+portNumber)
    if err != nil {
        fmt.Println("Failed to create tcp connection on port ", portNumber)
    }
    defer listner.Close()

    for {
        conn, err := listner.Accept()
        if err != nil {
            fmt.Println("Failed to connect")
        }
        
        fmt.Println("Connection successful with: ", conn.RemoteAddr())
        
        msgChannel := getLinesChannel(conn)
        for msgs := range msgChannel {
            fmt.Printf("%s\n", msgs)
        }
        fmt.Println("Closing connection, since channel is closed")
    }

}

func getLinesChannel(conn net.Conn) <-chan string {
    message := make(chan string)
    go func() {
        defer close(message)
        defer conn.Close()

        var b []byte = make([]byte, 8)

        var currLine string
        
        for n, err:= conn.Read(b); n > 0; {
            if err != nil {
                if currLine != "" {
                    message <- currLine
                }
                if errors.Is(err, io.EOF) {
                    break
                }
                fmt.Println(err.Error())
            }
            var splitString []string = strings.Split(string(b), "\n")
            currLine += splitString[0]
            for i := 1; i < len(splitString); i++ {
                message <-currLine
                currLine = splitString[i]
            }
            b = make([]byte, 8)
            n, _ = conn.Read(b)
        }
        message <- currLine
    } ()
    return message
}