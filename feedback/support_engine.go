package feedback

import (
	"database/sql"
	"strings"
)

type SupportTicketResult struct {
	Category       string
	UrgencyScore   int
	ShouldEscalate bool
	AutoReplyMsg   string
}

type SupportEngineService struct {
	db *sql.DB
}

func NewSupportEngineService(db *sql.DB) *SupportEngineService {
	return &SupportEngineService{db: db}
}

func (s *SupportEngineService) ProcessFeedback(phone, state, district, dialectCode, rawMessage string) *SupportTicketResult {
	lower := strings.ToLower(rawMessage)

	result := &SupportTicketResult{
		Category:       "GENERAL_FEEDBACK",
		UrgencyScore:   2,
		ShouldEscalate: false,
		AutoReplyMsg:   "आपके बहुमूल्य सुझाव के लिए धन्यवाद। हमने इसे दर्ज कर लिया है।",
	}

	if strings.Contains(lower, "पैसे कट गए") || strings.Contains(lower, "रुपये कट") || strings.Contains(lower, "फ्रॉड") || strings.Contains(lower, "काम नहीं कर रहा") {
		result.Category = "CRITICAL_COMPLAINT"
		result.UrgencyScore = 5
		result.ShouldEscalate = true
		result.AutoReplyMsg = "आदरणीय अभिभावक, आपकी समस्या को हमारी प्राथमिक तकनीकी टीम को भेज दिया गया है। हम 15 मिनट में समाधान करेंगे।"
	} else if strings.Contains(lower, "टेम") || strings.Contains(lower, "टाइम") || strings.Contains(lower, "समय") || strings.Contains(lower, "कटाई") {
		result.Category = "TIMING_SUGGESTION"
		result.UrgencyScore = 2
		result.AutoReplyMsg = "सुझाव के लिए धन्यवाद! हमने आपके बच्चे के दैनिक अभ्यास का समय आपकी सुविधानुसार अपडेट कर लिया है।"
	} else if strings.Contains(lower, "भारी") || strings.Contains(lower, "ज्यादा") || strings.Contains(lower, "बोझ") || strings.Contains(lower, "थक") {
		result.Category = "BURDEN_ISSUE"
		result.UrgencyScore = 3
		result.AutoReplyMsg = "हम समझ सकते हैं। बच्चे की गति के अनुसार आज के अभ्यास को केवल 10 मिनट के हल्के कार्य में बदल दिया गया है।"
	}

	s.db.Exec(`INSERT INTO parent_feedback_tickets 
		(student_phone, state, district, detected_dialect, raw_parent_message, sentiment_category, urgency_score, should_escalate) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		phone, state, district, dialectCode, rawMessage, result.Category, result.UrgencyScore, result.ShouldEscalate)

	return result
}
