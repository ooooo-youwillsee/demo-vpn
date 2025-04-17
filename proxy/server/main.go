package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":9080")
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 9080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("New client connected")

		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	httpConn, err := net.Dial("tcp", "127.0.0.1:7080")
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer httpConn.Close()

	go func() {
		_, err := io.Copy(httpConn, conn)
		if err != nil {
			fmt.Println("Error writing to client:", err)
			return
		}
	}()

	_, err = io.Copy(conn, httpConn)
	if err != nil {
		fmt.Println("Error writing to client:", err)
		return
	}
}
