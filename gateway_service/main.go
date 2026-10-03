package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"anant-abhyas/router"
	// _ "github.com/lib/pq"
)

var PricingPlans = map[int]struct {
	Name      string
	Price     float64
	Category  string
	IsPremium bool
}{
	1:  {"2-Day Demo", 9.0, "DEMO", false},
	2:  {"5-Day Demo", 17.0, "DEMO", false},
	3:  {"10-Day Demo", 29.0, "DEMO", false},
	4:  {"Basic Plan (1 Child)", 399.0, "CORE", false},
	5:  {"Pro Plan (1 Child)", 699.0, "CORE", true},
	6:  {"Family Pack (2 Children)", 1099.0, "CORE", true},
	7:  {"Unlimited Family (3 Children)", 1499.0, "CORE", true},
	8:  {"Competition Pack", 499.0, "SKILL", false},
	9:  {"Language: Indian", 299.0, "SKILL", false},
	10: {"Language: Global + Accent", 399.0, "SKILL", false},
}

const WebsiteLink = "https://www.anantabhyas.com"

type WhatsAppEngine struct {
	db *sql.DB
}

func init() {
	router.RegisterModule(&WhatsAppEngine{})
}

func (we *WhatsAppEngine) RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	we.db = db
	mux.HandleFunc("/api/v1/whatsapp/webhook", we.handleWebhook)
	mux.HandleFunc("/api/v1/billing/auto-debit-webhook", we.handleAutoDebitWebhook) // ऑटो-डेबिट और पेमेंट फेलियर हैंडलर
}

type WhatsAppWebhookPayload struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []struct {
					From string `json:"from"`
					Type string `json:"type"`
					Text struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

func (we *WhatsAppEngine) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		verifyToken := r.URL.Query().Get("hub.verify_token")
		expectedToken := os.Getenv("WHATSAPP_VERIFY_TOKEN")
		if verifyToken == expectedToken {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(r.URL.Query().Get("hub.challenge")))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload WhatsAppWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				phone := msg.From
				inputValue := ""
				if msg.Type == "text" {
					inputValue = msg.Text.Body
				}

				if phone != "" && inputValue != "" {
					go we.processUserMessage(phone, inputValue)
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// =====================================================================
// ऑटो-डेबिट और पेमेंट गेटवे का वेबहुक (सफलता या असफलता का निर्णय यहाँ होगा)
// =====================================================================
type BillingWebhookPayload struct {
	Phone         string  `json:"phone"`
	Status        string  `json:"status"` // "SUCCESS" या "FAILED"
	ChargedAmount float64 `json:"charged_amount"`
}

func (we *WhatsAppEngine) handleAutoDebitWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload BillingWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var subscriptionMonths int
	var childName string
	err := we.db.QueryRow(`SELECT subscription_months, COALESCE(child_name, 'विद्यार्थी') FROM users WHERE phone = $1`, payload.Phone).Scan(&subscriptionMonths, &childName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if payload.Status == "SUCCESS" {
		// ऑटो-डेबिट सफल हुआ: महीने का काउंटर बढ़ाएं
		nextMonthCount := subscriptionMonths + 1
		rewardMsg := ""

		// रिवॉर्ड चेक: चौथा (3), आठवां (7), और बारहवां (11) महीना पूरा होने पर 25% छूट
		if nextMonthCount == 3 || nextMonthCount == 7 || nextMonthCount == 11 {
			rewardMsg = "\n\n🎁 *बधाई हो!* कंपनी के रिवॉर्ड सिस्टम के तहत इस महीने आपको **25% की विशेष छूट** दी गई है!"
		}

		_, _ = we.db.Exec(`UPDATE users SET subscription_months = $1, onboarding_state = 'STATE_ACTIVE_STUDY' WHERE phone = $2`, nextMonthCount, payload.Phone)

		// WhatsApp पर सिर्फ सफलता और रिवॉर्ड की बधाई भेजें
		successMsg := fmt.Sprintf("✅ नमस्ते %s जी!\nइस महीने का ऑटो-डेबिट भुगतान (₹%.2f) सफलतापूर्वक हो गया है।%s\n\n📚 पढ़ाई जारी रखने के लिए यहाँ क्लिक करें:\n🔗 %s/login", childName, payload.ChargedAmount, rewardMsg, WebsiteLink)
		sendWhatsApp(payload.Phone, successMsg)

	} else if payload.Status == "FAILED" {
		// ऑटो-डेबिट फेल हो गया: काउंटर को रीसेट (0) करें क्योंकि नियम के मुताबिक गैप आने पर नया कस्टमर माना जाएगा
		_, _ = we.db.Exec(`UPDATE users SET subscription_months = 0, onboarding_state = 'STATE_PAYMENT_PENDING' WHERE phone = $1`, payload.Phone)

		// अब WhatsApp पर मैन्युअल भुगतान (QR / Payment Link) का विकल्प भेजेंगे
		failureMsg := fmt.Sprintf("⚠️ नमस्ते %s जी, आपका इस महीने का ऑटो-डेबिट असफल रहा है।\n\n👉 कृपया नीचे दिए गए लिंक से मैन्युअल भुगतान करें ताकि आपकी पढ़ाई न रुके:\n🔗 %s/pay?phone=%s", childName, WebsiteLink, payload.Phone)
		sendWhatsApp(payload.Phone, failureMsg)
	}

	w.WriteHeader(http.StatusOK)
}

func (we *WhatsAppEngine) processUserMessage(phone, inputValue string) {
	textLower := strings.ToLower(strings.TrimSpace(inputValue))

	var currentState, childName string
	err := we.db.QueryRow(`
		SELECT COALESCE(onboarding_state, 'NEW'), COALESCE(child_name, '') 
		FROM users WHERE phone = $1`, phone).Scan(&currentState, &childName)

	if err == sql.ErrNoRows {
		_, _ = we.db.Exec("INSERT INTO users (phone, onboarding_state, subscription_months) VALUES ($1, 'STATE_CHILD_NAME', 0)", phone)
		sendWhatsApp(phone, "🙏 *नमस्ते! 'अनंत अभ्यास' में आपका स्वागत है।*\n👉 1. आपके बच्चे का क्या नाम है?")
		return
	}

	if currentState == "STATE_ACTIVE_STUDY" {
		if textLower == "exam" || textLower == "hobby" || textLower == "camp" {
			sendWhatsApp(phone, fmt.Sprintf("👋 नमस्ते %s जी!\n\n👉 अभ्यास और एग्जाम मोड के लिए यहाँ क्लिक करें:\n🔗 %s/login", childName, WebsiteLink))
			return
		}
		sendWhatsApp(phone, fmt.Sprintf("📚 क्लासरूम लिंक: %s/login", WebsiteLink))
		return
	}

	switch currentState {
	case "STATE_CHILD_NAME":
		if textLower == "" {
			sendWhatsApp(phone, "⚠️ कृपया अपने बच्चे का सही नाम लिखकर भेजें।")
			return
		}
		_, _ = we.db.Exec("UPDATE users SET child_name = $1, onboarding_state = 'STATE_LANGUAGE' WHERE phone = $2", inputValue, phone)
		sendWhatsApp(phone, "🌐 28 भाषाओं में से किस भाषा में पढ़ाई करना चाहते हैं?")

	case "STATE_LANGUAGE":
		_, _ = we.db.Exec("UPDATE users SET study_language = $1, onboarding_state = 'STATE_CHOOSE_PLAN' WHERE phone = $2", inputValue, phone)
		planMenu := `🎉 प्रोफाइल बन गई है! अपनी जरूरत के अनुसार 1 से 10 तक का पैक नंबर चुनें:
1. 2-Day Demo (₹9) | 4. Basic (₹399) | 5. Pro (₹699) | 6. Family (₹1099)`
		sendWhatsApp(phone, planMenu)

	case "STATE_CHOOSE_PLAN":
		choice, err := strconv.Atoi(inputValue)
		if err != nil || choice < 1 || choice > 10 {
			sendWhatsApp(phone, "⚠️ कृपया 1 से 10 के बीच सही नंबर चुनें।")
			return
		}

		plan := PricingPlans[choice]
		
		// पहली बार जब यूजर प्लान चुनता है, तो ऑटो-डेबिट सेटअप का लिंक या निर्देश दिया जाता है
		_, _ = we.db.Exec("UPDATE users SET onboarding_state = 'STATE_PAYMENT_PENDING', pending_plan_id = $1 WHERE phone = $2", choice, phone)
		sendWhatsApp(phone, fmt.Sprintf("✅ %s चुने गए हैं।\n👉 ऑटो-डेबिट (Auto-Debit) सेटअप पूरा करने के लिए यहाँ क्लिक करें: %s/setup-autopay?plan=%d", plan.Name, WebsiteLink, choice))

	case "STATE_PAYMENT_PENDING":
		sendWhatsApp(phone, fmt.Sprintf("⏳ आपका ऑटो-डेबिट सेटअप पेंडिंग है। प्रक्रिया पूरी करने के लिए वेबसाइट पर जाएँ: %s/login", WebsiteLink))

	default:
		_, _ = we.db.Exec("UPDATE users SET onboarding_state = 'STATE_CHILD_NAME' WHERE phone = $1", phone)
		sendWhatsApp(phone, "🙏 क्षमा करें, कृपया फिर से शुरू करें: आपके बच्चे का क्या नाम है?")
	}
}

func sendWhatsApp(toPhone, msgText string) {
	// यहाँ रेंडर एनवायरनमेंट टोकन से चलने वाला आपका WhatsApp भेजने का फिक्स लॉजिक रहेगा
	log.Printf("[WHATSAPP NOTIFICATION to %s]: %s", toPhone, msgText)
}
