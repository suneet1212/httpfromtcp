package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const address string = "localhost:42069"

func main() {
    udpAddr, err := net.ResolveUDPAddr("udp", address)
    if err != nil {
        fmt.Println("Failed to crpeate udp connection at this address: ", address)
    }

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Failed to connect")
	}
	defer udpConn.Close()
	reader := bufio.NewReader(os.Stdin)

    for {
        fmt.Print(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("failed to read string from stdin")
		}

		_, err = udpConn.Write([]byte(line))
		if err != nil {
			fmt.Println("Failed to write to udp", err)
		}
		
    }

}
