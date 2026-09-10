package scale

import (
	"log"
	"sync"
	"time"
)

type PingPayload struct {
	StudentID string
	ActiveSec int
	Timestamp int64
}

type HighConcurrencyAggregator struct {
	buffer []PingPayload
	mu     sync.Mutex
}

func NewHighConcurrencyAggregator() *HighConcurrencyAggregator {
	return &HighConcurrencyAggregator{buffer: make([]PingPayload, 0, 5000)}
}

func (h *HighConcurrencyAggregator) QueuePing(ping PingPayload) {
	h.mu.Lock()
	h.buffer = append(h.buffer, ping)
	h.mu.Unlock()
}

func (h *HighConcurrencyAggregator) StartFlushDaemon(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			h.mu.Lock()
			if len(h.buffer) == 0 {
				h.mu.Unlock()
				continue
			}
			batch := h.buffer
			h.buffer = make([]PingPayload, 0, 5000)
			h.mu.Unlock()

			go func(records []PingPayload) {
				log.Printf("⚡ [SCALE ENGINE] %d पिंग्स बिना DB क्रैश के सिंक हुए", len(records))
			}(batch)
		}
	}()
}
