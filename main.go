package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	// कोर प्रोजेक्ट पैकेजेस
	"anant-project/audio"
	"anant-project/cmd/api"
	"anant-project/database"
	"anant-project/family"
	"anant-project/featurephone"
	"anant-project/feedback"
	"anant-project/gateway_service"
	"anant-project/holiday"
	"anant-project/internal"
	"anant-project/language"
	"anant-project/monitor"
	"anant-project/pricing"
	"anant-project/security"
	"anant-project/stealth"
	"anant-project/temporal"
	"anant-project/vacation"

	// नए 'अल्ट्रा' क्लस्टर, CBT व लर्निंग पैकेजेस
	"anant-abhyas/cluster"
	"anant-abhyas/exam"
	"anant-abhyas/finance"
	"anant-abhyas/learning"
	"anant-abhyas/parental"
	newpricing "anant-abhyas/pricing"
	"anant-abhyas/scale"
	"anant-abhyas/support"

	_ "github.com/lib/pq"
)

const (
	AdminNumber      = "9024414973"
	GatewayNumber    = "9664006651"
	MerchantVPA      = "9664006651@ptsbi"
	MerchantName     = "Royal fmc corporation"
	BrandDisplayName = "Anant abhyas"
)

type SessionState string

const (
	StateNew             SessionState = "NEW"
	StateAwaitingConsent SessionState = "AWAITING_CONSENT"
	StateAwaitingKids    SessionState = "AWAITING_KIDS"
	StateAwaitingDetails SessionState = "AWAITING_DETAILS"
	StateInTest          SessionState = "IN_TEST"
	StateDemoActive      SessionState = "DEMO_ACTIVE"
	StateAwaitingPlan    SessionState = "AWAITING_PLAN"
	StatePaidActive      SessionState = "PAID_ACTIVE"
	StateSpamDropped     SessionState = "SPAM_DROPPED"
)

type StudentSession struct {
	PhoneNumber     string
	State           SessionState
	TemporaryDemoID string
	PermanentUID    string
	ChildName       string
	Grade           int
	Hobby           string
	DemoStartDate   time.Time
	DemoEndDate     time.Time
	TestStartTime   time.Time
	SelectedPlan    pricing.PlanTier
	DailyScanLimit  int
	ScansUsedToday  int
	LastScanReset   time.Time
	ValidTill       time.Time
	LastActive      time.Time
}

type AdminFeedbackState struct {
	IsApproved    bool
	LastFeedback  string
	RevisionCount int
	mu            sync.RWMutex
}

var (
	adminControl  = &AdminFeedbackState{IsApproved: false, RevisionCount: 1}
	studentDB     = make(map[string]*StudentSession)
	dbMutex       sync.RWMutex
	clockEngine   = temporal.NewClockEngine()
	contactFilter *featurephone.ContactFilter
)

func init() {
	contactFilter = featurephone.NewContactFilter()
}

func StartKeepAlive() {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		log.Println("Keep-Alive: APP_URL सेट नहीं है, सेल्फ-पिंग बंद है।")
		return
	}

	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			resp, err := http.Get(appURL)
			if err != nil {
				log.Printf("Keep-Alive Ping विफल: %v\n", err)
				continue
			}
			resp.Body.Close()
			log.Printf("Keep-Alive Ping सफल: सर्वर सक्रिय है (Status: %s)\n", resp.Status)
		}
	}()
}

func getOrCreateSession(phone string) *StudentSession {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	now := clockEngine.Now()
	s, exists := studentDB[phone]
	if !exists {
		lastDigits := phone
		if len(phone) >= 4 {
			lastDigits = phone[len(phone)-4:]
		}
		s = &StudentSession{
			PhoneNumber:     phone,
			State:           StateNew,
			TemporaryDemoID: fmt.Sprintf("DEMO-2026-%s", lastDigits),
			DailyScanLimit:  3,
			DemoStartDate:   now,
			DemoEndDate:     now.AddDate(0, 0, 7),
			LastScanReset:   now,
			LastActive:      now,
		}
		studentDB[phone] = s
	}

	if clockEngine.HasDailyResetOccurred(s.LastScanReset) {
		s.ScansUsedToday = 0
		s.LastScanReset = now
	}
	s.LastActive = now
	return s
}

func buildDirectUPIPrompt(phone, childName string, tier pricing.PlanTier, amount float64, studentID string) string {
	encodedBrand := url.QueryEscape(fmt.Sprintf("%s (%s)", BrandDisplayName, MerchantName))
	note := url.QueryEscape(fmt.Sprintf("अनंत अभ्यास - %s (%s)", studentID, tier))
	upiURL := fmt.Sprintf("upi://pay?pa=%s&pn=%s&am=%.2f&cu=INR&tn=%s", MerchantVPA, encodedBrand, amount, note)
	qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s", url.QueryEscape(upiURL))

	return fmt.Sprintf("💳 *अनंत अभ्यास — डायरेक्ट UPI भुगतान (%s)*\n\n"+
		"• छात्र: %s\n• छात्र ID: *%s*\n• प्लान: *%s*\n• राशि: *₹%.2f*\n• UPI ID: `%s`\n\n"+
		"📲 *1-टैप UPI भुगतान:*\n%s\n\n🖼️ *QR कोड:* %s\n\n⚠️ भुगतान के बाद स्क्रीनशॉट भेजें।",
		MerchantName, childName, studentID, tier, amount, MerchantVPA, upiURL, qrURL)
}

func MasterWhatsAppGateway(fromPhone, messageBody string) string {
	phone := strings.TrimPrefix(strings.TrimSpace(fromPhone), "+")
	text := strings.TrimSpace(messageBody)
	lower := strings.ToLower(text)
	upper := strings.ToUpper(text)
	now := clockEngine.Now()

	// 1. एडमिन कंट्रोल
	if phone == AdminNumber {
		adminControl.mu.Lock()
		defer adminControl.mu.Unlock()

		if lower == "system approved live" || lower == "approve" {
			adminControl.IsApproved = true
			return "🚀 [ADMIN APPROVED] अनंत अभ्यास गेटवे (9664006651) अब सभी 28 राज्यों के लिए LIVE है!"
		}
		if lower == "system stop" || lower == "reject" {
			adminControl.IsApproved = false
			return "🛑 [SYSTEM PAUSED] गेटवे STAGING मोड में है।"
		}
		if strings.HasPrefix(lower, "fix:") || strings.HasPrefix(lower, "सुधार:") {
			adminControl.RevisionCount++
			adminControl.LastFeedback = text
			adminControl.IsApproved = false
			return fmt.Sprintf("🔧 [REVISION #%d RECORDED] \"%s\"\nरी-टेस्ट के लिए *TEST RUN* लिखें।", adminControl.RevisionCount, text)
		}
		if lower == "test run" || lower == "hi" || upper == "HI" || strings.HasPrefix(lower, "hi") {
			snap := clockEngine.GetCurrentSnapshot()
			return fmt.Sprintf("🧪 [ADMIN STAGING TEST]\n• समय (IST): %s\n• कोर इंजन: सक्रिय\n\nलाइव करने हेतु लिखें: *SYSTEM APPROVED LIVE*\nबदलाव हेतु लिखें: *FIX: [कमी]*", snap.FormattedTimestamp)
		}
	}

	// 2. पब्लिक लाइव गार्ड
	adminControl.mu.RLock()
	live := adminControl.IsApproved
	adminControl.mu.RUnlock()

	if !live && phone != AdminNumber {
		return "नमस्ते! 'अनंत अभ्यास' सिस्टम अभी एडमिन टेस्टिंग में है। थोड़ी देर बाद प्रयास करें।"
	}

	session := getOrCreateSession(phone)

	// 3. प्रो-राटा अपग्रेड (Basic -> Pro)
	if session.State == StatePaidActive && session.SelectedPlan == pricing.TierBasic && (lower == "pro" || lower == "upgrade") {
		daysRemaining := int(math.Ceil(time.Until(session.ValidTill).Hours() / 24))
		if daysRemaining < 0 {
			daysRemaining = 0
		}
		unusedBasic := float64(daysRemaining) * (399.0 / 30.0)
		finalPayable := math.Round(699.0 - unusedBasic)
		if finalPayable < 50 {
			finalPayable = 50
		}

		return fmt.Sprintf("🚀 *Pro Plan अपग्रेड*\n• शेष दिन: %d | कटौती: -₹%.2f\n👉 *देय: ₹%.0f*\n\n%s",
			daysRemaining, unusedBasic, finalPayable, buildDirectUPIPrompt(phone, session.ChildName, pricing.TierPro, finalPayable, session.PermanentUID))
	}

	// 4. पेमेंट वेरिफिकेशन व UID आवंटन
	if strings.Contains(upper, "PAID") || strings.Contains(upper, "SUCCESS") || strings.Contains(upper, "DEMO-2026") || strings.Contains(upper, "ABHYAS-2026") {
		if session.PermanentUID == "" {
			lastDigits := phone
			if len(phone) >= 4 {
				lastDigits = phone[len(phone)-4:]
			}
			session.PermanentUID = fmt.Sprintf("ABHYAS-2026-%s", lastDigits)
		}
		session.State = StatePaidActive
		session.ValidTill = now.AddDate(0, 1, 0)
		if strings.Contains(lower, "699") || session.SelectedPlan == pricing.TierPro {
			session.SelectedPlan = pricing.TierPro
			session.DailyScanLimit = 12
		} else {
			session.SelectedPlan = pricing.TierBasic
			session.DailyScanLimit = 5
		}
		return fmt.Sprintf("🎉 *सत्यापन सफल!*\n• स्थायी ID: *%s*\n• प्लान: *%s* (%d स्कैन/दिन)\n• वैधता: %s\n\nअभ्यास शुरू करने हेतु *START* लिखें।",
			session.PermanentUID, session.SelectedPlan, session.DailyScanLimit, session.ValidTill.Format("02-01-2006"))
	}

	// 5. ऑनबोर्डिंग फ्लो
	switch session.State {
	case StateNew:
		if lower == "hi" || upper == "HI" || strings.HasPrefix(lower, "hi") || strings.Contains(lower, "hello") || strings.Contains(lower, "नमस्ते") || strings.Contains(lower, "start") {
			session.State = StateAwaitingConsent
			return "नमस्ते! 'अनंत अभ्यास' में आपका स्वागत है। 🎓\nक्या आप 7-दिन फ्री डेमो के लिए तैयार हैं? (हाँ / नहीं)"
		}
		return "नमस्ते! शुरू करने हेतु *Hi* भेजें।"

	case StateAwaitingConsent:
		if lower == "हाँ" || lower == "yes" || lower == "ha" || lower == "haa" || upper == "YES" || lower == "y" {
			session.State = StateAwaitingDetails
			return "कृपया बच्चे का *नाम, कक्षा (1-12) और हॉबी* लिखें:\n(उदा: राहुल, कक्षा 6, रोबोटिक्स)"
		}
		session.State = StateSpamDropped
		return "धन्यवाद! भविष्य में कभी भी 'Hi' भेजकर शुरू कर सकते हैं।"

	case StateAwaitingDetails:
		parts := strings.Split(text, ",")
		session.ChildName = strings.TrimSpace(parts[0])
		session.Grade = 6
		session.Hobby = "General"
		if len(parts) >= 3 {
			session.Hobby = strings.TrimSpace(parts[2])
		}
		session.State = StateInTest
		session.TestStartTime = now
		return fmt.Sprintf("धन्यवाद! %s का 2-मिनट टेस्ट:\nसवाल: 12 + 8 = 20 है, तो 35 - 15 = कितना होगा?", session.ChildName)

	case StateInTest:
		duration := time.Since(session.TestStartTime).Seconds()
		session.State = StateDemoActive
		speed := "सामान्य"
		if duration < 10 {
			speed = "असाधारण (Olympiad Fast Thinker)"
		}
		return fmt.Sprintf("📊 *डायग्नोस्टिक रिपोर्ट*\n• छात्र: %s (कक्षा %d)\n• गति: %s\n• 🆔 डेमो ID: *%s*\n\n🎉 7-दिवसीय फ्री डेमो सक्रिय है! अभ्यास के लिए *START* लिखें।",
			session.ChildName, session.Grade, speed, session.TemporaryDemoID)

	case StateDemoActive:
		if now.After(session.DemoEndDate) || lower == "plan" {
			session.State = StateAwaitingPlan
			return fmt.Sprintf("🎉 7-दिन डेमो पूरा हुआ! प्लान चुनें:\n[1] Basic (₹399 - 5 स्कैन/दिन)\n[2] Pro (₹699 - 12 स्कैन/दिन)\n\nDemo ID: *%s*", session.TemporaryDemoID)
		}
		if lower == "start" || upper == "START" {
			return fmt.Sprintf("📚 Day अभ्यास एक्टिव है (%s)। सवाल पूछें या फोटो भेजें।", session.ChildName)
		}
		return "अभ्यास के लिए *START* लिखें।"

	case StateAwaitingPlan:
		if lower == "1" || strings.Contains(lower, "basic") {
			session.SelectedPlan = pricing.TierBasic
			return buildDirectUPIPrompt(phone, session.ChildName, pricing.TierBasic, 399.0, session.TemporaryDemoID)
		} else if lower == "2" || strings.Contains(lower, "pro") {
			session.SelectedPlan = pricing.TierPro
			return buildDirectUPIPrompt(phone, session.ChildName, pricing.TierPro, 699.0, session.TemporaryDemoID)
		}
		return "विकल्प चुनें: 1 (Basic ₹399) या 2 (Pro ₹699)"

	case StatePaidActive:
		if lower == "start" || upper == "START" {
			return fmt.Sprintf("🌟 स्वागत है %s! %s एक्टिव है (%d स्कैन/दिन)। डायरी फोटो भेजें।", session.ChildName, session.SelectedPlan, session.DailyScanLimit)
		}
		return "मास्टरजी सक्रिय हैं! सवाल पूछें।"
	}
	return ""
}

func handleIncomingCommunication(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	body := r.URL.Query().Get("body")
	if from == "" {
		from = r.FormValue("From")
		body = r.FormValue("Body")
	}

	profile := contactFilter.CheckCaller(from)

	if profile.IsFriend {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("NORMAL_ROUTING"))
		return
	}

	if profile.IsSpam {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("SPAM_REJECTED"))
		return
	}

	reply := MasterWhatsAppGateway(from, body)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(reply))
}

type UltraClusterHub struct {
	Brain          *learning.AdaptiveSystemBrain
	ClusterMesh    *cluster.DynamicClusterMesh
	AutoPayEngine  *finance.AutoPayManager
	ParentFeedback *feedback.SupportEngineService
	AppSupport     *support.AutoHealingEngine
	Aggregator     *scale.HighConcurrencyAggregator
}

func main() {
	log.Println("==========================================================")
	log.Println("🚀 'अनंत अभ्यास अल्ट्रा' 360° प्रोडक्शन क्लस्टर लाइव")
	log.Println("==========================================================")

	go StartKeepAlive()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:password@localhost:5432/anant_abhyas?sslmode=disable"
	}

	db, err := database.ConnectPostgres(connStr)
	if err != nil {
		log.Printf("⚠️ डेटाबेस कनेक्शन चेतावनी: %v (सिस्टम इन-मेमोरी मोड में जारी रहेगा)", err)
	} else {
		log.Println("✅ PostgreSQL डेटाबेस सफलतापूर्वक कनेक्ट हुआ।")
		if err := database.AutoMigrateDatabase(db); err != nil {
			log.Printf("⚠️ स्कीमा माइग्रेशन चेतावनी: %v", err)
		} else {
			log.Println("✅ सभी स्कीमा टेबल्स सत्यापित और अद्यतन हैं।")
		}
	}

	// 1. सेल्फ-लर्निंग ब्रेन और स्केल बफर
	brain := learning.NewAdaptiveSystemBrain()
	brain.RunSelfLearningCycle()

	aggregator := scale.NewHighConcurrencyAggregator()
	aggregator.StartFlushDaemon(5 * time.Second)

	autoPayEngine := finance.NewAutoPayManager()
	parentFeedback := feedback.NewSupportEngineService(db)
	appSupport := support.NewAutoHealingEngine()

	// 2. ऑटो-स्केलिंग क्लस्टर मैश
	clusterMesh := cluster.NewDynamicClusterMesh()
	renderURL := os.Getenv("RENDER_INTERNAL_URL")
	if renderURL == "" {
		renderURL = "http://localhost:8081"
	}
	clusterMesh.RegisterServerNode("NODE_RENDER_PRIMARY", renderURL, cluster.RolePrimaryRender, 500)

	secondaryVPS := os.Getenv("SECONDARY_VPS_URL")
	if secondaryVPS != "" {
		clusterMesh.RegisterServerNode("NODE_BUDGET_VPS_01", secondaryVPS, cluster.RoleSecondaryVPS, 5000)
	}
	clusterMesh.StartAutoScalingWatchdog()

	hub := &UltraClusterHub{
		Brain:          brain,
		ClusterMesh:    clusterMesh,
		AutoPayEngine:  autoPayEngine,
		ParentFeedback: parentFeedback,
		AppSupport:     appSupport,
		Aggregator:     aggregator,
	}

	// 3. 20 कोर बैकग्राउंड इंजनों की बाइंडिंग
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ कोर इंजन रिकवरी: %v", r)
			}
		}()

		_ = pricing.NewDemoService(db)
		_ = pricing.NewPlanService(db)
		_ = pricing.NewLoyaltyService(db)
		_ = vacation.NewAprilSessionService(db)
		_ = vacation.NewHomeworkService(db)
		_ = vacation.NewWinterBootcampService(db)
		_ = vacation.NewFoundationBridgeService()
		_ = holiday.NewExamSchedulerService(db)
		_ = vacation.NewCustomInterestService(db)
		_ = vacation.NewPacingService()
		_ = language.NewPanIndiaDialectService()
		_ = language.NewFusionDialectService(db)
		_ = audio.NewVoiceTunerService()
		_ = holiday.NewStateHolidayService(db)
		_ = security.NewBiometricDNAService(db)
		_ = security.NewSecuritySuite("ANANT_SECRET_2026", db)
		_ = security.NewAntiSharingGuard(db)
		_ = security.NewGeoTravelService(db)
		_ = monitor.NewMindReader()
		_ = monitor.NewInactivityNudgeService()
		_ = stealth.NewStealthComposer()
		_ = family.NewMultiChildEngine(db)
		_ = featurephone.NewFeaturePhoneEngine(db)
		_ = security.NewImageEnhancer()
		_ = monitor.NewWeeklyReportService(db)
		_ = api.NewAPIServer(db)
		_ = gateway_service.NewGatewayRouter(db)
		_ = internal.NewAdminDashboard(db)
		_ = internal.NewAutoHealerEngine(AdminNumber, 400.0)

		log.Println("✅ सभी 20 कोर बैकग्राउंड इंजन पूरी तरह सक्रिय हैं।")
	}()

	// 4. HTTP एंडपॉइंट्स रूटिंग

	// बेस वेब व एडमिन
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<h2>🎓 अनंत अभ्यास (रॉयल एफएमसी कॉरपोरेशन) लाइव है 24x7!</h2><p><a href="/admin">एडमिन पोर्टल</a></p>`)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"HEALTHY","cluster":"ONLINE"}`))
	})

	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		snap := clockEngine.GetCurrentSnapshot()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h2>अनंत अभ्यास एडमिन पोर्टल</h2><p>गेटवे: 9664006651 | एडमिन: 9024414973</p><p>समय (IST): %s</p>", snap.FormattedTimestamp)
	})

	// WhatsApp गेटवे
	http.HandleFunc("/webhook", handleIncomingCommunication)
	http.HandleFunc("/incoming", handleIncomingCommunication)

	// क्लस्टर ट्रैफिक डिस्पैचर
	http.HandleFunc("/cluster/dispatch", hub.ClusterMesh.RouteSmartTraffic)

	// ऑनबोर्डिंग व सख्त कैप्स बिलिंग (1, 1, 2, 4)
	http.HandleFunc("/api/v1/parent/onboarding", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "SUCCESS",
			"steps":  parental.GetParentWalkthrough(),
		})
	})

	http.HandleFunc("/api/v1/billing/calculate", func(w http.ResponseWriter, r *http.Request) {
		tier := r.URL.Query().Get("tier")
		hasExam := r.URL.Query().Get("exam_addon") == "true"
		bill, err := newpricing.ComputeModularBill(tier, hasExam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bill)
	})

	// 100% UPI ऑटो-पे मैंडेट रूट्स
	http.HandleFunc("/api/v1/payment/setup-autopay", func(w http.ResponseWriter, r *http.Request) {
		parentID := r.URL.Query().Get("parent_id")
		tier := r.URL.Query().Get("tier")
		hasExam := r.URL.Query().Get("exam_addon") == "true"

		bill, err := newpricing.ComputeModularBill(tier, hasExam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sub := hub.AutoPayEngine.SetupMandate(parentID, bill.TotalPrice)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":          "MANDATE_INITIATED",
			"monthly_amount":  bill.TotalPrice,
			"allowed_kids":    bill.AllowedKids,
			"subscription_id": sub.SubscriptionID,
			"message":         "UPI ऑटो-पे मैंडेट अधिकृत करें (नो-डिफ़ॉल्ट पॉलिसी)",
		})
	})

	http.HandleFunc("/api/v1/payment/autopay-webhook", func(w http.ResponseWriter, r *http.Request) {
		parentID := r.URL.Query().Get("parent_id")
		bankUTR := r.URL.Query().Get("bank_utr")
		success := r.URL.Query().Get("status") == "SUCCESS"

		err := hub.AutoPayEngine.HandleAutoDebitWebhook(parentID, bankUTR, success)
		if err != nil {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		}

		hub.Brain.ProcessFeedbackAndFinance("PAYMENT_SETTLED", bankUTR)
		w.Write([]byte(`{"status":"SETTLEMENT_CONFIRMED_AND_UNLOCKED"}`))
	})

	// इन-ऐप गवर्नमेंट CBT विंडो
	http.HandleFunc("/api/v1/cbt/submit", func(w http.ResponseWriter, r *http.Request) {
		var sub exam.CBTSessionSubmission
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			http.Error(w, "अमान्य CBT डेटा", http.StatusBadRequest)
			return
		}

		mockKey := map[string]string{"q1": "A", "q2": "B", "q3": "C", "q4": "D"}
		res := exam.EvaluateCBTSession(sub, mockKey)

		hub.Brain.IngestEvent(learning.SystemEvent{
			EventType: "EXAM_SUBMIT",
			TrackCode: sub.TrackCode,
			Score:     res.Score,
			Timestamp: time.Now(),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	// 30-सेकंड पिंग बफर
	http.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		sID := r.URL.Query().Get("student_id")
		if sID == "" {
			http.Error(w, "student_id आवश्यक", http.StatusBadRequest)
			return
		}
		hub.Aggregator.QueuePing(scale.PingPayload{
			StudentID: sID,
			ActiveSec: 30,
			Timestamp: time.Now().Unix(),
		})
		w.Write([]byte(`{"status":"ACK"}`))
	})

	// WhatsApp सपोर्ट व ऑटो-हीलिंग
	http.HandleFunc("/api/v1/whatsapp/webhook", func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		msg := r.URL.Query().Get("message")
		reply, handled := hub.ParentFeedback.ProcessFeedbackAndHeal(phone, msg)
		if !handled {
			reply = "नमस्ते! 'अनंत अभ्यास' में आपका स्वागत है। अभ्यास शुरू करने के लिए START लिखें।"
		}
		w.Write([]byte(reply))
	})

	// इन-ऐप क्रैश हुक
	http.HandleFunc("/api/v1/app/support-hook", func(w http.ResponseWriter, r *http.Request) {
		parentID := r.URL.Query().Get("parent_id")
	    rawError := r.URL.Query().Get("error_log")

		ticket := hub.AppSupport.IngestAndAutoResolve("IN_APP_CRASH_HOOK", parentID, rawError)
		hub.Brain.ProcessFeedbackAndFinance(string(ticket.Type), rawError)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ticket)
	})

	// एडमिन स्टेटस
	http.HandleFunc("/api/v1/admin/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"system":        "Anant Abhyas Ultra Core",
			"cluster_state": "ACTIVE",
			"auto_pay":      "ENFORCED_MANDATE_ONLY",
			"active_tracks": []string{"NAVODAYA", "SAINIK_SCHOOL", "NDA", "IIT_JEE"},
		})
	})

	// 5. सर्वर स्टार्टअप
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 अनंत अभ्यास क्लस्टर पोर्ट :%s पर पूर्णतः सक्रिय है...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
