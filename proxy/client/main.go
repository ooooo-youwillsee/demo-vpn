package main

import (
	"fmt"
	"github.com/songgao/water"
	"log"
	"os/exec"
)

func main() {
	// 配置 TUN 设备
	config := water.Config{
		DeviceType: water.TUN,
	}

	// 创建 TUN 接口
	tun, err := water.New(config)
	if err != nil {
		log.Fatalf("Failed to create TUN device: %v", err)
	}
	defer tun.Close()

	fmt.Printf("TUN device %s created.\n", tun.Name())

	err = setTunIpForMac(tun.Name(), "10.1.1.1", "10.1.1.2", "255.255.255.0")
	if err != nil {
		log.Fatalf("Failed to set TUN IP address: %v", err)
	}

	//serverConn, err := net.Dial("tcp", "127.0.0.1:9080")
	//if err != nil {
	//	fmt.Println("Error dialing:", err)
	//	return
	//}
	//defer serverConn.Close()
	//
	//_, err = serverConn.Write([]byte("hello server"))
	//if err != nil {
	//	fmt.Println("Error writing to server:", err)
	//	return
	//}
	//
	//buf := make([]byte, 1024)
	//n, err := serverConn.Read(buf)
	//if err != nil {
	//	fmt.Println("Error reading from server:", err)
	//	return
	//}
	//if msg := string(buf[:n]); msg != "hello client" {
	//	fmt.Println("receive server request error!!!")
	//	return
	//}

	// 读取 TUN 设备的数据包
	buffer := make([]byte, 1024*1024)
	for {
		n, err := tun.Read(buffer)
		if err != nil {
			log.Printf("Error reading from TUN device: %v", err)
			continue
		}
		fmt.Printf("Received a packet of length %d\n", n)
		fmt.Println(string(buffer[:n]))
	}

	//go func() {
	//	_, err := io.Copy(serverConn, tun)
	//	if err != nil {
	//		fmt.Println("tun -> serverConn, Error writing to client:", err)
	//		return
	//	}
	//}()
	//
	//go func() {
	//	_, err := io.Copy(tun, serverConn)
	//	if err != nil {
	//		fmt.Println("serverConn -> tun, Error writing to client:", err)
	//		return
	//	}
	//}()
	//
	// 处理系统信号，用于优雅退出
	//sigs := make(chan os.Signal, 1)
	//signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	//<-sigs
	//fmt.Println("Received termination signal, exiting...")
}

func setTunIpForMac(tunName, localIp, remoteIp, netmask string) error {
	// 使用 ifconfig 命令设置 IP 地址
	cmd := exec.Command("ifconfig", tunName, localIp, remoteIp, "netmask", netmask)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set IP address: %v, output: %s", err, string(output))
	}
	return nil
}

func setTunRoute(tunName, destNet, gateway string) error {
	// 使用 route 命令设置路由
	cmd := exec.Command("route", "add", "-net", destNet, "gw", gateway, tunName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set route: %v, output: %s", err, string(output))
	}
	return nil
}
