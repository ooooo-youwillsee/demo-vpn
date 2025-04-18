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
