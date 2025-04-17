package main

import (
	"demo-network/internalpenetration"
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
	log.Println("Client is listening on port 8080")

	httpConn := getHttpConn()
	serverConn := getServerConn()
	internalpenetration.CopyOnConn(httpConn.(*net.TCPConn), serverConn.(*net.TCPConn))

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, os.Kill)
	<-stopChan
}

func getHttpConn() net.Conn {
	httpConn, err := net.Dial("tcp", "127.0.0.1:7080")
	if err != nil {
		defer httpConn.Close()
		log.Println("Error dialing:", err)
		return nil
	}
	return httpConn
}
func getServerConn() net.Conn {
	serverConn, err := net.Dial("tcp", "127.0.0.1:9080")
	if err != nil {
		defer serverConn.Close()
		log.Println("Error dialing:", err)
		return nil
	}

	_, err = serverConn.Write([]byte("hello server"))
	if err != nil {
		log.Println("Error writing to server:", err)
		return nil
	}

	buf := make([]byte, 1024)
	n, err := serverConn.Read(buf)
	if err != nil {
		log.Println("Error reading from server:", err)
		return nil
	}
	if msg := string(buf[:n]); msg != "hello client" {
		log.Println("receive server request error!!!")
		return nil
	}
	return serverConn
}
