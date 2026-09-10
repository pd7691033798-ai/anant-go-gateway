package monitor

import (
	"errors"
	"time"
)

type AntiCheatGuard struct {
	LastQuestionServedAt time.Time
	MinAnswerSeconds     int
	Strikes              int
}

func NewAntiCheatGuard() *AntiCheatGuard {
	return &AntiCheatGuard{
		LastQuestionServedAt: time.Now(),
		MinAnswerSeconds:     3, // 3-सेकंड मिनिमम टाइमर (नो तुक्का)
		Strikes:              0,
	}
}

func (g *AntiCheatGuard) RecordQuestionServed() {
	g.LastQuestionServedAt = time.Now()
}

func (g *AntiCheatGuard) ValidateAttempt() (bool, error) {
	elapsed := time.Since(g.LastQuestionServedAt).Seconds()
	if elapsed < float64(g.MinAnswerSeconds) {
		g.Strikes++
		if g.Strikes >= 3 {
			return false, errors.New("CRITICAL_STRIKE_LOCK: 3 बार तुक्का मारा गया। पैरेंट पिन आवश्यक है")
		}
		return false, errors.New("INVALID_ATTEMPT: 3 सेकंड से पहले जवाब नहीं दे सकते")
	}
	return true, nil
}
