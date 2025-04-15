package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080")

	serverConn, err := net.Dial("tcp", "127.0.0.1:9080")
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer serverConn.Close()

	_, err = serverConn.Write([]byte("hello server"))
	if err != nil {
		fmt.Println("Error writing to server:", err)
		return
	}

	buf := make([]byte, 1024)
	n, err := serverConn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from server:", err)
		return
	}
	if msg := string(buf[:n]); msg != "hello client" {
		fmt.Println("receive server request error!!!")
		return
	}

	httpConn, err := net.Dial("tcp", "127.0.0.1:7080")
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer httpConn.Close()

	go func() {
		_, err := io.Copy(httpConn, serverConn)
		if err != nil {
			fmt.Println("Error writing to client:", err)
			return
		}
	}()
	go func() {
		_, err := io.Copy(serverConn, httpConn)
		if err != nil {
			fmt.Println("Error writing to client:", err)
			return
		}
	}()

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, os.Kill)
	<-stopChan
}
