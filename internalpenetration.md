## 内网穿透原理

> 内网穿透本质上是端口映射，和 NAT 很像，但是 NAT 是通过交换机来实现的，而内网穿透是通过内网代理来实现的。

主要角色：

1. 服务端：监听访问端请求，将请求转发给客户端
2. 客户端：负责主动和服务端，应用端建立连接，充当服务器和应用端之间的桥梁。
3. 应用端：负责接收客户端转发的请求，真实处理请求。

![img.png](img.png)

请求流程：

1. 启动 app 程序，并监听端口7080。
2. 启动 server 程序，并监听端口9080。
3. 启动 client 程序，与app，server建立连接。
4. client程序：负责将**server请求**复制到**app**，**app响应**复制到**server**。
5. server程序：负责将**app请求**复制到**client**，**client响应**复制到**app**。
6. 示例程序，没有处理**连接异常时重试，重连等**问题，过一段时间需要重启client。

示例代码：

app(一个简单的http程序)：

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

client：

```go 
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
```

server：

```go
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

```

util.go

```go
package internalpenetration

import (
	"io"
	"log"
	"net"
	"sync"
)

// CopyOnConn 负责dst和src的数据复制
func CopyOnConn(dst *net.TCPConn, src *net.TCPConn) {
	defer dst.Close()
	defer src.Close()
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer dst.CloseWrite()
		_, err := io.Copy(dst, src)
		if err != nil {
			log.Printf("%s->%s Error writing to client: %v \n", src.LocalAddr(), dst.LocalAddr(), err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		defer src.CloseWrite()
		_, err := io.Copy(src, dst)
		if err != nil {
			log.Printf("%s->%s Error writing to client: %v \n", dst.LocalAddr(), src.LocalAddr(), err)
			return
		}
	}()
	wg.Wait()
}
```


