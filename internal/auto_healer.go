package internal

import (
	"log"
	"net/http"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type AutoHealerEngine struct {
	AdminPhone      string
	MaxRAMThreshold float64
	CrashesSaved    int64
	mu              sync.RWMutex
}

func NewAutoHealerEngine(adminPhone string, maxRAM_MB float64) *AutoHealerEngine {
	healer := &AutoHealerEngine{
		AdminPhone:      adminPhone,
		MaxRAMThreshold: maxRAM_MB,
	}
	go healer.StartWatchdog(30 * time.Second)
	return healer
}

func (a *AutoHealerEngine) RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.mu.Lock()
				a.CrashesSaved++
				a.mu.Unlock()
				log.Printf("🚨 [CRASH PREVENTED] पैनिक: %v", err)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("⚠️ सिस्टम में समस्या आई थी, ऑटो-हीलिंग ने इसे संभाल लिया है।"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *AutoHealerEngine) StartWatchdog(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		currentRAM := float64(m.Alloc) / 1024 / 1024
		if currentRAM > a.MaxRAMThreshold {
			runtime.GC()
			debug.FreeOSMemory()
		}
	}
}
