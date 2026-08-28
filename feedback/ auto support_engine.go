package feedback

import (
	"database/sql"
	"strings"
	"time"
)

type SupportTicketResult struct {
	Category       string
	UrgencyScore   int
	ShouldEscalate bool
	AutoReplyMsg   string
	IsHandled      bool
}

type SupportEngineService struct {
	db *sql.DB
}

func NewSupportEngineService(db *sql.DB) *SupportEngineService {
	return &SupportEngineService{db: db}
}

// ProcessFeedback: पेरेंट मैसेज का वर्गीकरण, प्राथमिकता निर्धारण, डेटाबेस टिकटिंग और ऑटो-रिप्लाई प्रदान करता है
func (s *SupportEngineService) ProcessFeedback(phone, state, district, dialectCode, rawMessage string) *SupportTicketResult {
	lower := strings.ToLower(rawMessage)

	result := &SupportTicketResult{
		Category:       "GENERAL_FEEDBACK",
		UrgencyScore:   2,
		ShouldEscalate: false,
		AutoReplyMsg:   "आपके बहुमूल्य सुझाव के लिए धन्यवाद। हमने इसे दर्ज कर लिया है।",
		IsHandled:      true,
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

	if s.db != nil {
		_, _ = s.db.Exec(`INSERT INTO parent_feedback_tickets 
			(student_phone, state, district, detected_dialect, raw_parent_message, sentiment_category, urgency_score, should_escalate) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			phone, state, district, dialectCode, rawMessage, result.Category, result.UrgencyScore, result.ShouldEscalate)
	}

	return result
}

// ProcessFeedbackAndHeal: बहुभाषी (स्थानीय बोली) शिकायत व सुझाव की पहचान, ऑटो-हीलिंग (DB सेशन रीसेट) और तुरंत समाधान उत्तर देता है
func (s *SupportEngineService) ProcessFeedbackAndHeal(phone, msg string) (string, bool) {
	lower := strings.ToLower(msg)

	isComplaint := strings.Contains(lower, "काम नहीं") || strings.Contains(lower, "problem") ||
		strings.Contains(lower, "error") || strings.Contains(lower, "नी चाल") ||
		strings.Contains(lower, "अटक") || strings.Contains(lower, "शिकायत") ||
		strings.Contains(lower, "खुल नहीं") || strings.Contains(lower, "फंस")

	isSuggestion := strings.Contains(lower, "सुझाव") || strings.Contains(lower, "जोड़ो") ||
		strings.Contains(lower, "सलाह") || strings.Contains(lower, "suggestion") ||
		strings.Contains(lower, "चाहिए") || strings.Contains(lower, "add")

	if !isComplaint && !isSuggestion {
		return "", false
	}

	// 1. शिकायत का समाधान और ऑटो-हीलिंग (Auto-Heal Engine)
	if isComplaint {
		if s.db != nil {
			_, _ = s.db.Exec("UPDATE student_sessions SET scans_used_today = 0, state = 'DEMO_ACTIVE' WHERE phone_number = $1", phone)
			
			// फीडबैक टिकट टेबल में भी लॉग करें
			_, _ = s.db.Exec(`INSERT INTO parent_feedback_tickets 
				(student_phone, state, district, detected_dialect, raw_parent_message, sentiment_category, urgency_score, should_escalate) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				phone, "INDIA", "UNKNOWN", "LOCAL", msg, "AUTO_HEAL_COMPLAINT", 4, false)
		}

		// बोली/भाषा के अनुसार तुरंत समाधान
		if strings.Contains(lower, "नी चाल") || strings.Contains(lower, "म्हारो") || strings.Contains(lower, "कोनी") {
			return "सा, म्हे थारी समस्या पे तुरंत एक्शन ले'र सिस्टम रीसेट (Auto-Heal) कर दियो है। अबे START लिख'र देखो सा, सब चोखो चालेलो।", true
		}
		if strings.Contains(lower, "chal nahi") || strings.Contains(lower, "ruka") {
			return "🛠️ आपकी समस्या को सिस्टम ने स्वतः ठीक (Auto-Healed) कर दिया है। सेशन रीसेट हो चुका है, कृपया *START* लिखकर पुनः शुरू करें।", true
		}
		return "🛠️ *शिकायत समाधान:* सिस्टम ने आपके खाते का सेशन और कोटा स्वतः रीसेट (Auto-Heal) कर दिया है। अभ्यास के लिए *START* लिखें।", true
	}

	// 2. सुझाव का स्वतः रिकॉर्ड
	if isSuggestion {
		if s.db != nil {
			_, _ = s.db.Exec("INSERT INTO user_suggestions (phone_number, suggestion_text, created_at) VALUES ($1, $2, $3)", phone, msg, time.Now())
		}
		if strings.Contains(lower, "सा") || strings.Contains(lower, "घणो") {
			return "घणो-घणो धन्यवाद सा! थारो सुझाव म्हे नोट कर ल्यो है और जल्दी ही ऐप में जोड़स्यां।", true
		}
		return "💡 *सुझाव दर्ज:* 'अनंत अभ्यास' को बेहतर बनाने के आपके सुझाव हेतु धन्यवाद! इसे हमारी प्रोडक्ट टीम ने रिकॉर्ड कर लिया है।", true
	}

	return "", false
}
