package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const PORT = "localhost:8080"
func main() {
	// file, err := os.Open("messages.txt")
	udpAddr, err := net.ResolveUDPAddr("udp", PORT)
	if err != nil {
		log.Fatal("Could not connect to the address")
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal("Failed to connect to udp")
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')

		_, err = conn.Write([]byte(input))
		if err != nil {
			fmt.Println("Error: " + err.Error())
		}
	}

}