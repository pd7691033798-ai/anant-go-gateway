package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"anant-abhyas/router"
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
}

func (we *WhatsAppEngine) handleWebhook(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (we *WhatsAppEngine) processUserMessage(phone, inputType, inputValue string) {
	textLower := strings.ToLower(strings.TrimSpace(inputValue))

	var currentState, activePlan, childName string
	var subscriptionMonths int
	err := we.db.QueryRow(`
		SELECT onboarding_state, active_plan, child_name, subscription_months 
		FROM users WHERE phone = $1`, phone).Scan(&currentState, &activePlan, &childName, &subscriptionMonths)

	if err == sql.ErrNoRows {
		we.db.Exec("INSERT INTO users (phone, onboarding_state) VALUES ($1, 'STATE_CHILD_NAME')", phone)
		sendWhatsApp(phone, "🙏 *नमस्ते! 'अनंत अभ्यास' में आपका स्वागत है।*\n👉 1. आपके बच्चे का क्या नाम है?")
		return
	}

	if currentState == "STATE_ACTIVE_STUDY" {
		if textLower == "exam" || textLower == "hobby" || textLower == "camp" || inputType == "image" {
			redirectMsg := fmt.Sprintf("👋 नमस्ते %s जी!\n\n👉 अभ्यास और एग्जाम मोड के लिए यहाँ क्लिक करें:\n🔗 %s/login", childName, WebsiteLink)
			sendWhatsApp(phone, redirectMsg)
			return
		}
		sendWhatsApp(phone, fmt.Sprintf("📚 क्लासरूम लिंक: %s", WebsiteLink))
		return
	}

	switch currentState {
	case "STATE_CHILD_NAME":
		we.db.Exec("UPDATE users SET child_name = $1, onboarding_state = 'STATE_LANGUAGE' WHERE phone = $2", inputValue, phone)
		sendWhatsApp(phone, "🌐 28 भाषाओं में से किस भाषा में पढ़ाई करना चाहते हैं?")

	case "STATE_LANGUAGE":
		we.db.Exec("UPDATE users SET study_language = $1, onboarding_state = 'STATE_CHOOSE_PLAN' WHERE phone = $2", inputValue, phone)
		sendWhatsApp(phone, "🎉 प्रोफाइल बन गई है! 1 से 10 तक का पैक नंबर चुनें:")

	case "STATE_CHOOSE_PLAN":
		choice, err := strconv.Atoi(inputValue)
		if err != nil || choice < 1 || choice > 10 {
			sendWhatsApp(phone, "⚠️ कृपया 1 से 10 के बीच सही नंबर चुनें।")
			return
		}

		plan := PricingPlans[choice]
		finalPrice := plan.Price

		// 25% लॉयल्टी डिस्काउंट (चौथे, आठवें और बारहवें महीने के लिए)
		if subscriptionMonths == 3 || subscriptionMonths == 7 || subscriptionMonths == 11 {
			discount := finalPrice * 0.25
			finalPrice = finalPrice - discount
			sendWhatsApp(phone, fmt.Sprintf("🎁 25%% लॉयल्टी डिस्काउंट (₹%.2f छूट) लागू की गई है!", discount))
		}

		qrURL := getFeeQRCodeURL(phone, plan.Name, finalPrice)
		msg := fmt.Sprintf("✅ %s (₹%.2f)\n👉 QR कोड स्कैन करें। पेमेंट के बाद %s पर लॉगिन करें।", plan.Name, finalPrice, WebsiteLink)
		sendWhatsAppWithImage(phone, msg, qrURL)
		we.db.Exec("UPDATE users SET onboarding_state = 'STATE_PAYMENT_PENDING', pending_plan_id = $1 WHERE phone = $2", choice, phone)
	}
}

func sendWhatsApp(toPhone, msgText string) {}
func sendWhatsAppWithImage(toPhone, msgText, imgURL string) {}
func getFeeQRCodeURL(phone, planName string, amount float64) string { return "https://dummy-qr.com" }
