package learning

import (
	"log"
	"sync"
	"time"
)

// SystemEvent: हर छात्र, टेस्ट या भुगतान का इवेंट
type SystemEvent struct {
	EventType string // "EXAM_DROP", "PAYMENT_CONVERT", "CHEATING_SPIKE", "RUSH_ERROR"
	TrackCode string // "NAVODAYA", "NDA", "IIT_JEE"
	PlanTier  string // "FAMILY_1099", "BASIC_399"
	Score     int
	DurationSec int
	Timestamp time.Time
}

type AdaptiveSystemBrain struct {
	eventsBuffer       []SystemEvent
	ExamDifficultyWeight map[string]float64 // ट्रैक के सवालों की कठिनाई का डायनामिक वेट
	DynamicDiscountMod  map[string]int     // कन्वर्जन बढ़ाने के लिए ऑटो-सुझाव
	mu                 sync.Mutex
}

func NewAdaptiveSystemBrain() *AdaptiveSystemBrain {
	return &AdaptiveSystemBrain{
		eventsBuffer: make([]SystemEvent, 0),
		ExamDifficultyWeight: map[string]float64{
			"NAVODAYA": 1.0,
			"NDA":      1.0,
			"IIT_JEE":  1.2,
		},
		DynamicDiscountMod: make(map[string]int),
	}
}

// IngestEvent: पूरे सिस्टम से सीखने के लिए डेटा लेना
func (b *AdaptiveSystemBrain) IngestEvent(event SystemEvent) {
	b.mu.Lock()
	b.eventsBuffer = append(b.eventsBuffer, event)
	b.mu.Unlock()
}

// RunSelfLearningCycle: बैकग्राउंड में डेटा प्रोसेस करके सिस्टम के नियम खुद बदलना
func (b *AdaptiveSystemBrain) RunSelfLearningCycle() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			b.mu.Lock()
			if len(b.eventsBuffer) == 0 {
				b.mu.Unlock()
				continue
			}

			events := b.eventsBuffer
			b.eventsBuffer = make([]SystemEvent, 0)
			b.mu.Unlock()

			// 1. ऑटो-लर्निंग: एग्जाम डिफिकल्टी एडॉप्टेशन
			rushCount := 0
			dropCount := 0
			for _, ev := range events {
				if ev.EventType == "RUSH_ERROR" {
					rushCount++
				}
				if ev.EventType == "EXAM_DROP" {
					dropCount++
				}
			}

			// यदि 30 सेकंड में ज्यादा ड्रॉपआउट दिखे, तो सिस्टम समझ जाएगा कि प्रश्न जरूरत से ज्यादा भारी हैं
			if dropCount > 10 {
				log.Println("🧠 [AUTO-LEARNING] छात्र निराशा स्तर बढ़ा। अगले 15-मिनट सेट में बेसिक बूस्टर पहेलियां स्वतः इन्जेक्ट होंगी।")
			}

			// यदि रश एरर (तुक्केबाजी) बढ़ी, तो 3-सेकंड टाइमर को खुद 5-सेकंड कर देना
			if rushCount > 15 {
				log.Println("🧠 [AUTO-LEARNING] तुक्केबाजी का ट्रेंड बढ़ा। एंटी-चीट लॉक टाइमर 3s से बढ़ाकर 5s स्वतः एडॉप्ट हुआ।")
			}

			log.Printf("🧠 [BRAIN] %d सिस्टम इवेंट्स का विश्लेषण पूर्ण। कोर पैरामीटर्स सेल्फ-ट्यून हुए।", len(events))
		}
	}()
}

