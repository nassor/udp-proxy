package main

import (
	"crypto/rand"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	stopping := make(chan os.Signal, 1)
	signal.Notify(stopping, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:4404")
	checkError(err)

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err)

	workers := runtime.NumCPU() * 5
	for i := 0; i < workers; i++ {
		go generateLoad(conn)
	}

	<-stopping

}

func generateLoad(conn *net.UDPConn) {
	for {
		token := make([]byte, 1500)
		rand.Read(token)
		_, err := conn.Write(token)
		if err != nil {
			log.Println(err.Error())
			time.Sleep(5 * time.Second)
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}
