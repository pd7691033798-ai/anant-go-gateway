package middleware

import (
	"database/sql"
	"net/http"
	"strings"
)

// PlanValidationMiddleware यूजर के प्लान टियर और उसके दैनिक कोटे की जाँच करता है
func PlanValidationMiddleware(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. अनुरोध से यूजर का व्हाट्सएप नंबर या फोन प्राप्त करें
		userPhone := r.Header.Get("X-User-Phone")
		if userPhone == "" {
			http.Error(w, "अनधिकृत एक्सेस: फोन नंबर नहीं मिला", http.StatusUnauthorized)
			return
		}

		// 2. डेटाबेस से यूजर का वर्तमान प्लान टियर चेक करें
		var planTier string
		err := db.QueryRow("SELECT plan_tier FROM users WHERE phone = $1", userPhone).Scan(&planTier)
		if err != nil {
			http.Error(w, "यूजर डेटाबेस में नहीं मिला", http.StatusForbidden)
			return
		}

		// 3. प्लान के आधार पर अनुमतियाँ तय करें
		// यदि यूजर 'UNLIMITED_FAMILY', 'FAMILY', या 'PRO' में है, तो उसे सीधी अनुमति दें
		planTier = strings.ToUpper(planTier)
		if planTier == "UNLIMITED_FAMILY" || planTier == "FAMILY" || planTier == "PRO" {
			// असीमित एक्सेस, आगे बढ़ें
			next.ServeHTTP(w, r)
			return
		}

		// यदि यूजर 'DEMO' या 'BASIC' में है, तो दैनिक उपयोग (Daily Usage) का कोटा चेक करें
		if planTier == "DEMO" || planTier == "BASIC" {
			var scansUsed, qaUsed int
			err = db.QueryRow(
				"SELECT scans_used, qa_questions_used FROM daily_usage WHERE whatsapp_number = $1 AND usage_date = CURRENT_DATE", 
				userPhone,
			).Scan(&scansUsed, &qaUsed)

			// यदि आज का रिकॉर्ड नहीं है, तो नया दिन माना जाएगा (कोटा सुरक्षित)
			if err == nil {
				// उदाहरण के लिए: डेमो यूजर के लिए दैनिक सीमा 5 सवाल या स्कैन है
				if planTier == "DEMO" && (scansUsed >= 5 || qaUsed >= 5) {
					http.Error(w, "आपका दैनिक डेमो कोटा समाप्त हो गया है। कृपया प्लान अपग्रेड करें!", http.StatusPaymentRequired)
					return
				}
			}
		}

		// सब कुछ सही होने पर आगे के राउट पर जाने दें
		next.ServeHTTP(w, r)
	}
}
