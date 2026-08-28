package security

import (
	"strings"
	"sync"
	"time"
)

type SpamGuard struct {
	mu           sync.RWMutex
	attempts     map[string]int
	lastSeen     map[string]time.Time
	shadowBanned map[string]bool
	banExpiry    map[string]time.Time
	userDayCount map[string]int
}

func NewSpamGuard() *SpamGuard {
	return &SpamGuard{
		attempts:     make(map[string]int),
		lastSeen:     make(map[string]time.Time),
		shadowBanned: make(map[string]bool),
		banExpiry:    make(map[string]time.Time),
		userDayCount: make(map[string]int),
	}
}

// IsSpamOrBot बोट्स, तेज़ टाइपिंग और स्पैम संदेशों की जाँच करता है
func (g *SpamGuard) IsSpamOrBot(phone, message string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()

	// शैडो-बैन एक्सपायरी चेक (12 घंटे बाद स्वतः अनबैन)
	if g.shadowBanned[phone] {
		if now.Before(g.banExpiry[phone]) {
			return true
		}
		delete(g.shadowBanned, phone)
		delete(g.banExpiry, phone)
		g.attempts[phone] = 0
	}

	// 1-सेकंड से कम समय में रिस्पॉन्स (Bot/Script Detection)
	last, exists := g.lastSeen[phone]
	g.lastSeen[phone] = now
	if exists && now.Sub(last) < 1*time.Second {
		g.attempts[phone]++
		if g.attempts[phone] >= 3 {
			g.shadowBanned[phone] = true
			g.banExpiry[phone] = now.Add(12 * time.Hour)
			return true
		}
	}

	// कीवर्ड और पैटर्न फ़िल्टर
	lower := strings.ToLower(message)
	junkWords := []string{"http://", "https://", "free cash", "rummy", "betting", "lottery", "earn money"}
	for _, kw := range junkWords {
		if strings.Contains(lower, kw) {
			g.shadowBanned[phone] = true
			g.banExpiry[phone] = now.Add(24 * time.Hour)
			return true
		}
	}

	return false
}

// GetProgressiveScanLimit दिन के अनुसार 1, 2, या 3 स्कैन की सीमा तय करता है
func (g *SpamGuard) GetProgressiveScanLimit(dayNumber int) int {
	switch dayNumber {
	case 1:
		return 1 // Day 1: केवल 1 स्कैन (सत्यापन)
	case 2:
		return 2 // Day 2: 2 स्कैन
	default:
		return 3 // Day 3-7: 3 स्कैन/दिन
	}
}

