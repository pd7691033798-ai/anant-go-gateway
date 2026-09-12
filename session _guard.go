package monitor

import (
	"errors"
	"sync"
	"time"
)

type SessionState string

const (
	StateActive        SessionState = "ACTIVE"
	StateWarning5Min   SessionState = "WARNING_5M"
	StateGraceOvertime SessionState = "GRACE_OVERTIME"
	StateLocked        SessionState = "LOCKED"
)

type StudentSession struct {
	SessionID       string       `json:"session_id"`
	StudentPhone    string       `json:"student_phone"`
	StartedAt       time.Time    `json:"started_at"`
	ScheduledEndAt  time.Time    `json:"scheduled_end_at"`  // StartedAt + 60m
	HardGraceEndAt  time.Time    `json:"hard_grace_end_at"` // ScheduledEndAt + 60s
	CurrentState    SessionState `json:"current_state"`
	LastStrokeEvent time.Time    `json:"last_stroke_event"` // छात्र के अंतिम टच/पेन का समय
	IsSubmitted     bool         `json:"is_submitted"`
}

type SessionGuardService struct {
	mu       sync.RWMutex
	sessions map[string]*StudentSession
}

func NewSessionGuardService() *SessionGuardService {
	return &SessionGuardService{
		sessions: make(map[string]*StudentSession),
	}
}

// StartNewSession: 60 मिनट का नया अभ्यास सत्र शुरू करता है
func (s *SessionGuardService) StartNewSession(sessionID, phone string) *StudentSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	session := &StudentSession{
		SessionID:       sessionID,
		StudentPhone:    phone,
		StartedAt:       now,
		ScheduledEndAt:  now.Add(60 * time.Minute),
		HardGraceEndAt:  now.Add(61 * time.Minute), // 60 सेकंड का अनिवार्य ग्रेस
		CurrentState:    StateActive,
		LastStrokeEvent: now,
		IsSubmitted:     false,
	}

	s.sessions[sessionID] = session
	return session
}

// CheckSessionStatus: ऐप के प्रत्येक 10-सेकंड पिंग पर स्थिति की गणना करता है
func (s *SessionGuardService) CheckSessionStatus(sessionID string, hasActiveStroke bool) (*StudentSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.New("सत्र अमान्य है या समाप्त हो चुका है")
	}

	if sess.IsSubmitted {
		sess.CurrentState = StateLocked
		return sess, nil
	}

	now := time.Now()
	if hasActiveStroke {
		sess.LastStrokeEvent = now
	}

	// 1. यदि 61 मिनट (हार्ड ग्रेस) भी समाप्त हो चुके हैं
	if now.After(sess.HardGraceEndAt) {
		sess.CurrentState = StateLocked
		return sess, nil
	}

	// 2. यदि 60 मिनट पूरे हो चुके हैं (ग्रेस बफ़र निर्णय)
	if now.After(sess.ScheduledEndAt) {
		// यदि छात्र पिछले 90 सेकंड में सक्रिय था, तो 60 सेकंड का अतिरिक्त ग्रेस दें
		if time.Since(sess.LastStrokeEvent) < 90*time.Second {
			sess.CurrentState = StateGraceOvertime
			return sess, nil
		}
		sess.CurrentState = StateLocked
		return sess, nil
	}

	// 3. यदि 55 मिनट पूरे हो चुके हैं (5 मिनट की चेतावनी)
	warningThreshold := sess.StartedAt.Add(55 * time.Minute)
	if now.After(warningThreshold) {
		sess.CurrentState = StateWarning5Min
		return sess, nil
	}

	sess.CurrentState = StateActive
	return sess, nil
}

// SubmitLastQuestion: छात्र द्वारा अंतिम प्रश्न सबमिट करते ही तुरंत लॉक करना
func (s *SessionGuardService) SubmitLastQuestion(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("सत्र नहीं मिला")
	}

	sess.IsSubmitted = true
	sess.CurrentState = StateLocked
	return nil
}
