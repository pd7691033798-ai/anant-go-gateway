package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"

	"anant-abhyas/internal/admin"
	"anant-abhyas/internal/featurephone"
	"anant-abhyas/internal/gateway"
	"anant-abhyas/internal/onboarding"
	"anant-abhyas/internal/tier1"
	"anant-abhyas/internal/tier2"
	"anant-abhyas/internal/vault"
)

func main() {
	vaultManager := vault.NewProductionVaultManager("32_BYTE_SECRET_KEY_FOR_AES_PROD")
	vaultManager.VerifyExecutionIntegrity()

	connStr := "postgres://user:pass@localhost:5432/anant_abhyas_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ डेटाबेस कनेक्शन विफल: %v", err)
	}
	defer db.Close()

	onboardingEngine := onboarding.NewThreeStageFormEngine(db)
	dialectSwitcher := featurephone.NewDialectSwitcher(db)
	personalContacts := gateway.NewPersonalContactsManager(db)
	callGateway := gateway.NewCallOrchestrator(db, personalContacts)
	adminSearch := admin.NewKundaliSearchEngine(db)
	adminDash := admin.NewAdminDashboard(db)
	statsHandler := admin.NewAdminStatsHandler(db)
	tier1Suite := tier1.NewCoreLaunchSuite(db)
	tier2Suite := tier2.NewAsyncEngineSuite(db)

	// vCard 1-टैप कॉन्टैक्ट सेवर
	http.HandleFunc("/save", vaultManager.ServeVCardHandler)

	// 3-स्टेज ऑनबोर्डिंग WhatsApp वेबहुक
	http.HandleFunc("/api/v1/whatsapp/webhook", func(w http.ResponseWriter, r *http.Request) {
		from := r.URL.Query().Get("From")
		body := r.URL.Query().Get("Body")
		reply := onboardingEngine.ProcessUserInput(from, body)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(reply))
	})

	// टेलीफोनी गेटवे व पर्सनल कॉन्टैक्ट सिंक
	http.HandleFunc("/api/v1/gateway/incoming-call", callGateway.HandleMissCall)
	http.HandleFunc("/api/v1/gateway/sync-contacts", personalContacts.SyncPhonebookContacts)

	// डायनामिक भाषा स्विचिंग
	http.HandleFunc("/api/v1/session/switch-dialect", func(w http.ResponseWriter, r *http.Request) {
		uid := r.URL.Query().Get("uid")
		key := r.URL.Query().Get("key")
		newLang, prompt := dialectSwitcher.SwitchDialect(uid, key)
		json.NewEncoder(w).Encode(map[string]string{"dialect": newLang, "prompt": prompt})
	})

	// एडमिन डैशबोर्ड, कुंडली सर्च और लाइव आंकड़े
	http.HandleFunc("/admin", adminDash.ServeDashboardHTML)
	http.HandleFunc("/api/v1/admin/kundali", adminSearch.HandleKundaliSearch)
	http.HandleFunc("/api/v1/admin/stats", statsHandler.GetStatsJSON)
	http.HandleFunc("/api/v1/admin/active-count", func(w http.ResponseWriter, r *http.Request) {
		cnt, _ := tier2Suite.GetActiveStudentCount()
		json.NewEncoder(w).Encode(map[string]int{"active_students": cnt})
	})

	// पेमेंट QR जनरेटर
	http.HandleFunc("/api/v1/payment/qr", func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		plan := r.URL.Query().Get("plan")
		month, _ := strconv.Atoi(r.URL.Query().Get("month"))
		uri, amt := tier1Suite.GeneratePaymentQR(phone, plan, month)
		json.NewEncoder(w).Encode(map[string]interface{}{"qr_uri": uri, "amount": amt})
	})

	port := ":8080"
	log.Printf("🚀 अनंत अभ्यास बैकएंड पोर्ट %s पर 100%% सक्रिय है। गेटवे: %s", port, gateway.OfficialGatewayNumber)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("सर्वर त्रुटि: %v", err)
	}
}

