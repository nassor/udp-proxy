package proxy

import (
	"fmt"
	"sync/atomic"
	"time"
)

const outputChannels = uint16(2000)
const outputChannelBuffer = uint16(2000) // TODO: invented number
const updPacketSize = uint16(1492)       // TODO: MTU - UDP HEADER based

// processor will be able to:
// 1. flag the package to be discarded
// 2. inform with output channel the package belongs
// 3. modify the package
type processor interface {
	Process(payload [updPacketSize]byte) (discard bool, outputID uint16, modifiedPayload [updPacketSize]byte)
}

// TODO: STUN, Player to Location
type Proxy struct {
	// player X, Y ,Z are communicating using the channel A
	output [outputChannels]chan [updPacketSize]byte
	pp     processor

	stats Stats
}

type Stats struct {
	input    uint64
	output   uint64
	inflight int64
}

func New(poolSize int, processor processor) *Proxy {
	var output [outputChannels]chan [updPacketSize]byte
	for i := uint16(0); i < outputChannels; i++ {
		output[i] = make(chan [updPacketSize]byte, outputChannelBuffer)
	}

	return &Proxy{
		output: output,
		pp:     processor,
	}
}

func (p *Proxy) Add(payload [updPacketSize]byte) error {
	atomic.AddUint64(&p.stats.input, 1)
	atomic.AddInt64(&p.stats.inflight, 1)

	discard, outputID, modifiedPayload := p.pp.Process(payload)

	if discard {
		atomic.AddInt64(&p.stats.inflight, -1)
		return nil
	}

	// if the output is full throw it away
	if uint16(len(p.output[outputID])) == outputChannelBuffer {
		// TODO should be stats not log
		fmt.Printf("output channel full at id %d\n", outputID)
	}

	// Send to the correct output channel
	p.output[outputID] <- modifiedPayload

	return nil
}

func (p *Proxy) Start() {
	// for now just throwing all output away
	for i := uint16(0); i < outputChannels; i++ {
		go p.sendToOutput(p.output[i])
	}
}

// Stats  will print the amount of input and output
// instead terminal it should be set at a metric
func (p *Proxy) Stats() {
	input := atomic.LoadUint64(&p.stats.input)
	output := atomic.LoadUint64(&p.stats.output)
	atomic.StoreUint64(&p.stats.input, 0)
	atomic.StoreUint64(&p.stats.output, 0)
	inflight := atomic.LoadInt64(&p.stats.inflight)

	fmt.Printf("input: %d/sec\t|\toutput: %d/sec\t|\tinflight:%d\n", input, output, inflight)
}

// sendToOutput: for now just reading the channel and throwing everything away
func (p *Proxy) sendToOutput(output chan [updPacketSize]byte) {
	for range output {
		atomic.AddUint64(&p.stats.output, 1)
		atomic.AddInt64(&p.stats.inflight, -1)
	}
}

// Stop close all channel and wait for them to be empty
func (p *Proxy) Stop() {
	for i := uint16(0); i < outputChannels; i++ {
		if len(p.output[i]) != 0 {
			time.Sleep(time.Millisecond)
			i = i - 1
		}
	}
	for i := uint16(0); i < outputChannels; i++ {
		close(p.output[i])
	}
}
