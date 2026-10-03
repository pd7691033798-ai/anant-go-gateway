package voicegateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os" // [नया] Environment Variable के लिए
	"strings"
	"time"

	"anant-abhyas/router" // सेंट्रल रजिस्ट्री
)

type VoiceModule struct{}

func init() {
	router.RegisterModule(&VoiceModule{})
}

type IVRResponse struct {
	Response string `json:"response"`
}

type WebhookPayload struct {
	Token string `json:"token"`
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

func (v *VoiceModule) RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	// [सुधार] 10M स्केल ऑप्टिमाइज़ेशन: सर्वर स्टार्ट होते ही टोकन को मेमोरी में सेव कर लिया।
	// इससे 1 लाख रिक्वेस्ट आने पर भी बार-बार os.Getenv नहीं चलेगा, जिससे CPU बचेगा।
	expectedToken := os.Getenv("IVR_SECRET_TOKEN")
	if expectedToken == "" {
		// अगर तूने लाइव सर्वर पर टोकन डालना भूल गया, तो यह स्टार्ट होते ही टर्मिनल में लाल बत्ती जला देगा!
		log.Println("[CRITICAL WARNING] IVR_SECRET_TOKEN is NOT set in environment variables! IVR API will block all calls.")
	}

	mux.HandleFunc("/api/v1/voice/call-hook", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		// 1. HTTP Method सुरक्षा
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			sendCustomError(w, "Invalid Request Method", http.StatusMethodNotAllowed)
			return
		}

		var secretToken, phone, spokenName string

		if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			var payload WebhookPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
				secretToken = payload.Token
				phone = payload.Phone
				spokenName = payload.Name
			}
		} else {
			secretToken = r.FormValue("token")
			phone = r.FormValue("phone")
			spokenName = r.FormValue("name")
		}

		// 2. फौलादी ऑथेंटिकेशन (अब .env से मैच करेगा)
		// अगर expectedToken सर्वर पर सेट ही नहीं है, तो भी सिक्यूरिटी के लिए सब ब्लॉक रहेगा।
		if expectedToken == "" || secretToken != expectedToken {
			sendCustomError(w, "Unauthorized Access", http.StatusUnauthorized)
			return
		}

		cleanPhone := cleanPhoneNumber(phone)
		if cleanPhone == "" {
			sendJSONResponse(w, "नमस्ते! कृपया एक वैध 10-अंकीय मोबाइल नंबर से संपर्क करें।")
			return
		}

		// 3. डेटाबेस टाइमआउट
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		var existingName sql.NullString
		query := `SELECT name FROM parents_or_users WHERE phone = $1`
		err := db.QueryRowContext(ctx, query, cleanPhone).Scan(&existingName)

		var replyMessage string

		if err != nil && err != sql.ErrNoRows {
			log.Printf("[CRITICAL ERROR] Database query failed for %s: %v", cleanPhone, err)
			sendJSONResponse(w, "नमस्ते! अभी सर्वर पर काफी लोड है, कृपया 5 मिनट बाद कॉल करें।")
			return
		}

		// मौजूदा कस्टमर
		if err == nil {
			displayName := "अभिभावक"
			if existingName.Valid && strings.TrimSpace(existingName.String) != "" {
				displayName = existingName.String
			}
			
			replyMessage = "नमस्ते " + displayName + " जी! 'अनंत अभ्यास' में वापसी पर आपका स्वागत है। " +
				"आज के अभ्यास के लिए 1 दबाएं, अपने पिछले टेस्ट का रिजल्ट जानने के लिए 2 दबाएं।"
		} else {
			// नया कस्टमर
			if strings.TrimSpace(spokenName) == "" {
				spokenName = "अभिभावक"
			}

			now := time.Now().UTC()
			insertQuery := `
				INSERT INTO parents_or_users (phone, name, registered_at) 
				VALUES ($1, $2, $3)
				ON CONFLICT (phone) DO UPDATE SET name = EXCLUDED.name`
			
			_, dbErr := db.ExecContext(ctx, insertQuery, cleanPhone, spokenName, now)
			if dbErr != nil {
				log.Printf("[ERROR] Failed to insert new user %s: %v", cleanPhone, dbErr)
				sendJSONResponse(w, "नमस्ते! नेटवर्क में कुछ परेशानी है, कृपया थोड़ी देर में दोबारा प्रयास करें।")
				return
			}

			replyMessage = "नमस्ते " + spokenName + " जी! 'अनंत अभ्यास' लर्निंग प्लेटफॉर्म में आपका स्वागत है। " +
				"आपका नंबर हमारे सिस्टम में सुरक्षित रूप से दर्ज हो गया है, हमारी टीम जल्द ही आपसे संपर्क करेगी।"
		}

		sendJSONResponse(w, replyMessage)
	})
}

func sendCustomError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	sendJSONResponse(w, message)
}

func sendJSONResponse(w http.ResponseWriter, message string) {
	resp := IVRResponse{Response: message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] JSON encode failed: %v", err)
	}
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
	return ""
}
