package family

import (
	"database/sql"
	"fmt"
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

// 60 मिनट में 4 बच्चों के लिए 15-15 मिनट का अनुक्रमिक स्लॉट बनाना
func (m *MultiChildEngine) GenerateOneHourSchedule(parentPhone string, children []ChildSession) []ChildSession {
	now := time.Now()
	slotDuration := 15 * time.Minute
	schedule := make([]ChildSession, len(children))

	for i, child := range children {
		start := now.Add(time.Duration(i) * slotDuration)
		end := start.Add(slotDuration)
		child.SlotStartTime = start
		child.SlotEndTime = end
		schedule[i] = child
	}
	return schedule
}

func (m *MultiChildEngine) GetCurrentActiveChild(schedule []ChildSession) (*ChildSession, string) {
	now := time.Now()
	for _, child := range schedule {
		if now.After(child.SlotStartTime) && now.Before(child.SlotEndTime) {
			remainingMinutes := int(child.SlotEndTime.Sub(now).Minutes())
			msg := fmt.Sprintf("⏱️ अभी %s (कक्षा %d) का 15-मिनट अभ्यास स्लॉट सक्रिय है। शेष समय: %d मिनट।", child.Name, child.Grade, remainingMinutes)
			return &child, msg
		}
	}
	return nil, "आज का 1-घंटे का संयुक्त पारिवारिक अभ्यास सत्र पूर्ण हो चुका है।"
}
