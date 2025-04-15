package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

var (
	clientConn net.Conn
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
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		conn.Close()
		return
	}

	if msg := string(buf[:n]); msg == "hello server" {
		fmt.Println("receive client request")
		clientConn = conn
		_, err = conn.Write([]byte("hello client"))
	} else {
		fmt.Println("receive app request")
		if clientConn != nil {
			clientConn.Write(buf[:n])
			go func() {
				_, err2 := io.Copy(clientConn, conn)
				if err2 != nil {
					fmt.Println("Error writing to client:", err2)
					return
				}
			}()
			go func() {
				defer conn.Close()
				_, err2 := io.Copy(conn, clientConn)
				if err2 != nil {
					fmt.Println("Error writing to client:", err2)
					return
				}
			}()
		}
	}
}
