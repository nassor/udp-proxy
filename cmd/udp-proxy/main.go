package main

import (
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/nassor/udp-proxy/internal/proxy"
)

func main() {
	stopping := make(chan os.Signal, 1)
	signal.Notify(stopping, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	workers := runtime.NumCPU() * 10
	rp := proxy.RandomProcessor{}
	proxier := proxy.New(workers, rp)
	proxier.Start()

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:4404")
	checkError(err)
	conn, err := net.ListenUDP("udp", udpAddr)
	checkError(err)
	udpReceiver := proxy.NewReceiver(proxier, conn)
	for i := 0; i < workers; i++ {
		go udpReceiver.Listen()
	}

	// DEBUG
	//print stats every second on the console
	go func() {
		for {
			proxier.Stats()
			time.Sleep(time.Second)
		}
	}()

	// it allows http profiler
	go func() {
		http.ListenAndServe("127.0.0.1:3000", nil)
	}()
	// END: DEBUG

	log.Println("service started")

	<-stopping
	log.Println("stopping service...")
	udpReceiver.Stop()
	log.Println("udp receiver stopped...")
	proxier.Stop()
	log.Println("proxer stopped...")
	log.Println("service stopped.")

}

func checkError(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}
