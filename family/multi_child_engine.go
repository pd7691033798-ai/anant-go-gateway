package family

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ChildSession struct {
	ChildID       string
	Name          string
	Grade         int
	SlotStartTime time.Time
	SlotEndTime   time.Time
	IsCompleted   bool
}

type MultiChildEngine struct {
	db *sql.DB
}

func NewMultiChildEngine(db *sql.DB) *MultiChildEngine {
	return &MultiChildEngine{db: db}
}

// 📌 माता-पिता द्वारा चुने गए कुल समय और प्लान की सीमा (Family: अधिकतम 2, Unlimited Family: अधिकतम 3) के आधार पर डायनेमिक स्लॉट बनाना
func (m *MultiChildEngine) GenerateDynamicSchedule(parentPhone string, planTier string, totalMinutes int, children []ChildSession) ([]ChildSession, error) {
	numChildren := len(children)
	if numChildren == 0 {
		return nil, errors.New("कम से कम एक बच्चे की प्रोफाइल होना आवश्यक है")
	}

	cleanTier := strings.ToUpper(strings.TrimSpace(planTier))

	// 1. सख्त प्लान सीमा जाँच (Family में max 2, Unlimited Family में max 3)
	if cleanTier == "FAMILY" && numChildren > 2 {
		return nil, errors.New("अस्वीकृत: 'FAMILY' पैक में अधिकतम केवल 2 बच्चों की अनुमति है। 3 बच्चों के लिए कृपया 'UNLIMITED_FAMILY' पैक पर अपग्रेड करें।")
	}

	if cleanTier == "UNLIMITED_FAMILY" && numChildren > 3 {
		return nil, errors.New("अस्वीकृत: 'UNLIMITED_FAMILY' पैक में अधिकतम 3 बच्चों की सीमा तय की गई है।")
	}

	// यदि कोई अन्य या डेमो प्लान है और बच्चे 2 से अधिक हैं
	if cleanTier != "FAMILY" && cleanTier != "UNLIMITED_FAMILY" && numChildren > 2 {
		return nil, errors.New("अस्वीकृत: इस प्लान में बच्चों की संख्या की अनुमति नहीं है।")
	}

	// यदि कुल समय नहीं दिया गया है, तो डिफ़ॉल्ट 60 मिनट मान लें
	if totalMinutes <= 0 {
		totalMinutes = 60 
	}

	// 2. कुल समय को बच्चों की संख्या के अनुसार बराबर भागों में विभाजित करना
	slotMinutes := float64(totalMinutes) / float64(numChildren)
	slotDuration := time.Duration(slotMinutes * float64(time.Minute))

	now := time.Now()
	schedule := make([]ChildSession, numChildren)

	for i, child := range children {
		start := now.Add(time.Duration(i) * slotDuration)
		end := start.Add(slotDuration)
		child.SlotStartTime = start
		child.SlotEndTime = end
		schedule[i] = child
	}

	return schedule, nil
}

// ⏱ वर्तमान में सक्रिय बच्चे की जाँच करने का इंजन
func (m *MultiChildEngine) GetCurrentActiveChild(schedule []ChildSession) (*ChildSession, string) {
	now := time.Now()
	for _, child := range schedule {
		if (now.Equal(child.SlotStartTime) || now.After(child.SlotStartTime)) && now.Before(child.SlotEndTime) {
			remainingMinutes := int(child.SlotEndTime.Sub(now).Minutes())
			if remainingMinutes < 1 {
				remainingMinutes = 1
			}
			msg := fmt.Sprintf("⏱️ अभी %s (कक्षा %d) का अभ्यास स्लॉट सक्रिय है। शेष समय: %d मिनट।", child.Name, child.Grade, remainingMinutes)
			return &child, msg
		}
	}
	return nil, "आज का निर्धारित पारिवारिक अभ्यास सत्र पूर्ण हो चुका है।"
}
