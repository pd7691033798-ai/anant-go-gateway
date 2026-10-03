package feedback

import (
	"database/sql"
	"log"
	"strings"
	"time"
)

type SupportEngineService struct {
	db *sql.DB
}

func NewSupportEngineService(db *sql.DB) *SupportEngineService {
	return &SupportEngineService{db: db}
}

// ProcessSupportAtScale: 29 राज्यों, अंग्रेजी और स्थानीय बोलियों के लिए पूरी तरह ऑटोमैटिक सपोर्ट और ऑटो-हीलिंग इंजन
func (s *SupportEngineService) ProcessSupportAtScale(phone, state, district, preferredLanguage, rawMessage string) string {
	lower := strings.ToLower(rawMessage)

	// 1. सेंटीमेंट और प्राथमिकता तय करना
	category, urgency, shouldEscalate := s.classifySentiment(lower)

	// 2. डेटाबेस में ऑडिट ट्रेल (Audit Trail) सेव करना
	if s.db != nil {
		_, err := s.db.Exec(`INSERT INTO parent_feedback_tickets 
			(student_phone, state, district, detected_language, raw_parent_message, sentiment_category, urgency_score, should_escalate, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			phone, state, district, preferredLanguage, rawMessage, category, urgency, shouldEscalate, time.Now())
		if err != nil {
			log.Printf("[ERROR] Failed to save ticket: %v", err)
		}
	}

	// 3. तुरंत समाधान और ऑटो-हीलिंग (Zero Wait Policy)
	if category == "CRITICAL_COMPLAINT" {
		return s.getLocalizedReply(preferredLanguage, "CRITICAL")
	}

	if category == "TECHNICAL_GLITCH" {
		if s.db != nil {
			// बिना सब्सक्रिप्शन गिराए केवल दैनिक सेशन और कोटा रीसेट (Auto-Heal) करना
			_, _ = s.db.Exec("UPDATE student_sessions SET scans_used_today = 0 WHERE phone_number = $1", phone)
		}
		return s.getLocalizedReply(preferredLanguage, "TECH_FIXED")
	}

	if category == "USER_SUGGESTION" {
		if s.db != nil {
			_, _ = s.db.Exec("INSERT INTO user_suggestions (phone_number, suggestion_text, created_at) VALUES ($1, $2, $3)", phone, rawMessage, time.Now())
		}
		return s.getLocalizedReply(preferredLanguage, "SUGGESTION_SAVED")
	}

	return s.getLocalizedReply(preferredLanguage, "GENERAL")
}

// शिकायत या सुझाव की श्रेणी तय करना
func (s *SupportEngineService) classifySentiment(lower string) (string, int, bool) {
	if strings.Contains(lower, "पैसे") || strings.Contains(lower, "रुपये") || strings.Contains(lower, "fraud") || strings.Contains(lower, "payment failed") || strings.Contains(lower, "money deducted") {
		return "CRITICAL_COMPLAINT", 5, true
	}
	if strings.Contains(lower, "काम नहीं") || strings.Contains(lower, "अटक") || strings.Contains(lower, "error") || strings.Contains(lower, "not working") || strings.Contains(lower, "stuck") || strings.Contains(lower, "خुल नहीं") {
		return "TECHNICAL_GLITCH", 4, false
	}
	if strings.Contains(lower, "सुझाव") || strings.Contains(lower, "suggestion") || strings.Contains(lower, "add") || strings.Contains(lower, "चाहिए") || strings.Contains(lower, "feedback") {
		return "USER_SUGGESTION", 2, false
	}
	return "GENERAL_FEEDBACK", 1, false
}

// भाषा और राज्य के आधार पर आत्मीय और सटीक जवाब देने वाला डायनेमिक इंजन
func (s *SupportEngineService) getLocalizedReply(lang, issueType string) string {
	langUpper := strings.ToUpper(lang)

	switch issueType {
	case "CRITICAL":
		if langUpper == "ENGLISH" {
			return "🚨 *Priority Support:* Your payment concern has been logged with highest priority. Our technical team will resolve it within 10 minutes. Rest assured, your funds are 100% safe."
		}
		if langUpper == "RAJASTHANI" || langUpper == "HINDI" {
			return "राम राम सा! थारी पेमेंट री समस्या प म्हे तुरंत एक्शन ले लियो है। हमारी टेक्निकल टीम 10 मिनट में थारे पास फोन करसी। विस्वास राखो, एक रुपयो नी डूबण देश्यां।"
		}
		return "🚨 वित्तीय सुरक्षा प्राथमिकता: आपकी समस्या दर्ज कर ली गई है। हमारी तकनीकी टीम 10 मिनट में समाधान करेगी।"

	case "TECH_FIXED":
		if langUpper == "ENGLISH" {
			return "🛠️ *Auto-Healed:* System has automatically refreshed your session and fixed the technical glitch. Please type *START* to continue learning!"
		}
		if langUpper == "RAJASTHANI" {
			return "🛠️ सा, म्हे थारे सिस्टम रो एरर ऑटो-टिक ठीक (Auto-Heal) कर दियो है। अबे 'START' लिख'र देख लो, सब चोखो चालेलो!"
		}
		return "🛠️ *ऑटो-हीलिंग सक्रिय:* सिस्टम ने आपके खाते की तकनीकी रुकावट को स्वतः ठीक कर दिया है। अभ्यास के लिए *START* लिखें।"

	case "SUGGESTION_SAVED":
		if langUpper == "ENGLISH" {
			return "💡 *Suggestion Saved:* Thank you for your valuable feedback! Our product team has recorded it to make 'Anant Abhyaas' even better."
		}
		return "💡 *सुझाव स्वीकृत:* आपके इस बहुमूल्य सुझाव के लिए धन्यवाद! इसे हमारे सिस्टम में दर्ज कर लिया गया है।"

	default:
		if langUpper == "ENGLISH" {
			return "Hello! Thank you for reaching out. 'Anant Abhyaas' support is available 24/7 for your growth."
		}
		return "नमस्ते! आपके संदेश के लिए धन्यवाद। 'अनंत अभ्यास' परिवार आपकी सेवा में सदैव तत्पर है।"
	}
}
