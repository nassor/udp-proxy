package proxy

import (
	"math/rand"
)

type RandomProcessor struct {
}

func (p RandomProcessor) Process(payload [updPacketSize]byte) (bool, uint16, [updPacketSize]byte) {
	discard := false
	n := rand.Int31n(7) // simulating package discards
	if n == 1 {
		discard = true
	}

	outputID := uint16(rand.Int31n(int32(outputChannelBuffer)))

	payload[134] = 32
	payload[135] = 22
	payload[166] = 22

	return discard, outputID, payload
}
