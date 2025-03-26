package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:6666")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// goroutine1: listening server
	go func() {
		connScanner := bufio.NewScanner(conn)
		for connScanner.Scan() {
			fmt.Println(connScanner.Text())
			fmt.Printf("> ")
		}
		if err := connScanner.Err(); err != nil {
			log.Println("Error reading from server:", err)
		}
	}()

	// goroutine2: listening client
	stdinScanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("> ")
		if !stdinScanner.Scan() {
			break
		}
		text := stdinScanner.Text()
		fmt.Fprintf(conn, "%s\n", text) // send to server
	}

	fmt.Println("Client exiting.")
}
