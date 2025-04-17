package main

import (
	"demo-network/internalpenetration"
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
	log.Println("Server is listening on port 9080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Println("New client connected")
		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		defer conn.Close()
		log.Println("Error reading from connection:", err)
		return
	}

	if msg := string(buf[:n]); msg == "hello server" {
		log.Println("receive client request")
		_, _ = conn.Write([]byte("hello client"))
		clientConn = conn
	} else {
		log.Println("receive app request")
		handleAppRequest(buf[:n], conn)
	}
}

func handleAppRequest(buf []byte, conn net.Conn) {
	if clientConn == nil {
		return
	}
	_, _ = clientConn.Write(buf)
	internalpenetration.CopyOnConn(clientConn.(*net.TCPConn), conn.(*net.TCPConn))
}
