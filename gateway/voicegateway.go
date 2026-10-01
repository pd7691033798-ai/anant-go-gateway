package voicegateway

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"anant-abhyas/router" // सेंट्रल रजिस्ट्री पैकेज
)

type VoiceModule struct{}

// Go का init() फंक्शन स्वतः इस मॉड्यूल को सेंट्रल रजिस्ट्री में जोड़ देता है
func init() {
	router.RegisterModule(&VoiceModule{})
}

// यह फंक्शन सेंट्रल रजिस्ट्री के जरिए main.go से अपने आप राउट्स जोड़ लेगा
func (v *VoiceModule) RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/api/v1/voice/call-hook", func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		spokenName := r.URL.Query().Get("name") // आईवीआर या स्पीच-टू-टेक्स्ट से मिला नाम

		cleanPhone := cleanPhoneNumber(phone)
		if cleanPhone == "" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(`{"response": "नमस्ते! कृपया वैध नंबर से संपर्क करें।"}`))
			return
		}

		// 1. डेटाबेस में चेक करें कि क्या यह नंबर पहले से रजिस्टर्ड है या नहीं
		var existingName string
		query := `SELECT name FROM parents_or_users WHERE phone = $1`
		err := db.QueryRow(query, cleanPhone).Scan(&existingName)

		var replyMessage string

		// 2. यदि नंबर पहले से मौजूद है (पुराना/सम्मानित ग्राहक या अभिभावक)
		if err == nil && existingName != "" {
			replyMessage = fmt.Sprintf(
				"नमस्ते %s जी! 'अनंत अभ्यास' में आपका स्वागत है। आपके बेहतर अनुभव के लिए बताइए, आपको किस प्रकार की समस्या हुई? जो परेशानी आपको हुई, उसके लिए हमें बेहद खेद है। बताइए %s जी, हम आपकी क्या मदद कर सकते हैं?",
				existingName, existingName,
			)
		} else {
			// 3. यदि नया नंबर है, तो नाम सेव करें और नया आत्मीय स्वागत संदेश दें
			if spokenName == "" {
				spokenName = "अभिभावक"
			}

			insertQuery := `
				INSERT INTO parents_or_users (phone, name, registered_at) 
				VALUES ($1, $2, NOW())
				ON CONFLICT (phone) DO UPDATE SET name = EXCLUDED.name`
			_, _ = db.Exec(insertQuery, cleanPhone, spokenName)

			replyMessage = fmt.Sprintf(
				"नमस्ते %s जी! 'अनंत अभ्यास' लर्निंग प्लेटफॉर्म में आपका स्वागत है। इससे जुड़ने के लिए आपका धन्यवाद। यहाँ कक्षा 1 से 12 तक और प्रतियोगी परीक्षाओं की बेहतरीन तैयारी कराई जाती है। आपकी जानकारी दर्ज हो गई है, हमारी टीम जल्द ही आपसे संपर्क करेगी।",
				spokenName,
			)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(fmt.Sprintf(`{"response": "%s"}`, replyMessage)))
	})
}

func cleanPhoneNumber(val string) string {
	digits := ""
	for _, ch := range val {
		if ch >= '0' && ch <= '9' {
			digits += string(ch)
		}
	}
	if len(digits) >= 10 {
		return digits[len(digits)-10:]
	}
	return digits
}
