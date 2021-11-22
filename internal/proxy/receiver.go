package proxy

import (
	"log"
	"net"
	"sync/atomic"
	"time"
)

type proxier interface {
	Add(payload [updPacketSize]byte) error
}

type Receiver struct {
	running uint32
	conn    *net.UDPConn

	proxier proxier
}

func NewReceiver(proxier proxier, conn *net.UDPConn) *Receiver {
	return &Receiver{
		running: 1,
		proxier: proxier,
		conn:    conn,
	}
}

func (r *Receiver) Listen() {
	// Add to a sync.Pool if request is coming on multiple payloads sizes
	// (if needs to be a slice, not an array)
	buffer := [updPacketSize]byte{}

	for {
		if atomic.LoadUint32(&r.running) != 1 {
			return
		}

		_, _, err := r.conn.ReadFromUDP(buffer[:])
		if err != nil {
			log.Printf("when readubg UDP payload: %s\n", err)
			continue
		}

		if err = r.proxier.Add(buffer); err != nil {
			log.Printf("when sending to proxy: %s\n", err)
		}
	}
}

func (r *Receiver) Stop() {
	log.Println("stopped called")
	atomic.CompareAndSwapUint32(&r.running, 1, 0)
	time.Sleep(time.Second)
}
