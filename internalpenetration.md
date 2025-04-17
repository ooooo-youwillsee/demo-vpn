## 内网穿透原理

> 内网穿透本质上是端口映射，和 NAT 很像，但是 NAT 是通过交换机来实现的，而内网穿透是通过内网代理来实现的。

主要角色：

1. 服务端：监听访问端请求，将请求转发给客户端
2. 客户端：负责主动和服务端，应用端建立连接，充当服务器和应用端之间的桥梁。
3. 应用端：负责接收客户端转发的请求，真实处理请求。

![img.png](img.png)


示例代码：

app.go：
```go
package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)

	server := http.Server{
		Addr:    ":7080",
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		return
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World from APP"))
}

```

客户端：

```go 
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
			fmt.Println("serverConn -> httpConn, Error writing to client:", err)
			return
		}
	}()
	go func() {
		_, err := io.Copy(serverConn, httpConn)
		if err != nil {
			fmt.Println("httpConn -> serverConn, Error writing to client:", err)
			return
		}
	}()

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, os.Kill)
	<-stopChan
}
```


服务端：

```go
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
		handleAppRequest(buf[:n], conn)
	}
}

func handleAppRequest(buf []byte, conn net.Conn) {
	if clientConn == nil {
		return
	}
	_, _ = clientConn.Write(buf)
	go func() {
		_, err2 := io.Copy(clientConn, conn)
		if err2 != nil {
			fmt.Println("conn -> clientConn, Error writing to client:", err2)
			return
		}
	}()
	go func() {
		_, err2 := io.Copy(conn, clientConn)
		if err2 != nil {
			fmt.Println("clientConn -> conn, Error writing to client:", err2)
			return
		}
	}()
}

```


