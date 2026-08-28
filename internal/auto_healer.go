package internal

import (
	"context"
	"log"
	"net/http"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type AutoHealerEngine struct {
	AdminPhone      string
	MaxRAMThreshold float64 // मेगाबाइट्स (MB) में
	crashesSaved    int64
	mu              sync.RWMutex
	stopChan        chan struct{}
}

func NewAutoHealerEngine(adminPhone string, maxRAM_MB float64) *AutoHealerEngine {
	if maxRAM_MB <= 0 {
		maxRAM_MB = 512.0 // डिफ़ॉल्ट 512 MB सीमा
	}

	healer := &AutoHealerEngine{
		AdminPhone:      adminPhone,
		MaxRAMThreshold: maxRAM_MB,
		stopChan:        make(chan struct{}),
	}

	go healer.StartWatchdog(30 * time.Second)
	return healer
}

// RecoverMiddleware किसी भी अनपेक्षित पैनिक को रोककर सर्वर को चालू रखता है
func (a *AutoHealerEngine) RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.mu.Lock()
				a.crashesSaved++
				crashes := a.crashesSaved
				a.mu.Unlock()

				// विस्तृत एरर और स्टैक ट्रेस लॉगिंग
				stack := string(debug.Stack())
				log.Printf("🚨 [CRASH PREVENTED #%d] Path: %s | Panic: %v\nStack Trace:\n%s", crashes, r.URL.Path, err, stack)

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "सिस्टम में अस्थायी समस्या आई थी, ऑटो-हीलर ने स्थिति को संभाल लिया है।"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// StartWatchdog नियमित अंतराल पर RAM की जांच करता है और सीमा पार होने पर GC चलाता है
func (a *AutoHealerEngine) StartWatchdog(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			currentRAM := float64(m.Alloc) / 1024 / 1024

			if currentRAM > a.MaxRAMThreshold {
				log.Printf("⚠️ [RAM Watchdog] मेमोरी सीमा पार: %.2f MB / %.2f MB. गारबेज कलेक्शन शुरू...", currentRAM, a.MaxRAMThreshold)
				runtime.GC()
				debug.FreeOSMemory()

				var afterMem runtime.MemStats
				runtime.ReadMemStats(&afterMem)
				freedRAM := currentRAM - (float64(afterMem.Alloc) / 1024 / 1024)
				log.Printf("✅ [RAM Watchdog] मेमोरी साफ़ की गई। मुक्त हुई RAM: %.2f MB", freedRAM)
			}
		case <-a.stopChan:
			log.Println("🛑 [RAM Watchdog] वॉचडॉग सुरक्षित रूप से बंद हुआ।")
			return
		}
	}
}

// Stop वॉचडॉग को सुरक्षित रोकने के लिए
func (a *AutoHealerEngine) Stop() {
	close(a.stopChan)
}

// GetCrashesSaved थ्रेड-सेफ़ तरीके से कुल रोके गए क्रैश की संख्या देता है
func (a *AutoHealerEngine) GetCrashesSaved() int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.crashesSaved
}
