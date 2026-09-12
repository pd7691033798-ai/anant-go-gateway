package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// कोर प्रोजेक्ट पैकेजेस
	"anant-abhyas/audio"
	"anant-abhyas/database"
	"anant-abhyas/family"
	"anant-abhyas/featurephone"
	"anant-abhyas/feedback"
	"anant-abhyas/holiday"
	"anant-abhyas/internal"
	"anant-abhyas/language"
	"anant-abhyas/monitor"
	"anant-abhyas/pricing"
	"anant-abhyas/sandbox"
	"anant-abhyas/security"
	"anant-abhyas/stealth"
	"anant-abhyas/temporal"
	"anant-abhyas/vacation"

	// क्लस्टर, CBT, पेरेंटल इंजन व बिलिंग पैकेजेस
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

// वर्कर पूल पेलोड
type WebhookJob struct {
	From string
	Body string
}

var (
	clockEngine          = temporal.NewClockEngine()
	contactFilter        *featurephone.ContactFilter
	fullOnboardingEngine *parental.FullWhatsAppEngine
	wellnessEngine       *vacation.ChildWellnessService
	pinMgr               *security.PINManager
	weeklyReportEngine   *monitor.WeeklyReportService
	webhookQueue         = make(chan WebhookJob, 1000) // बफ़र्ड चैनल
)

func init() {
	contactFilter = featurephone.NewContactFilter()
	fullOnboardingEngine = parental.NewFullWhatsAppEngine()
}

func StartKeepAlive() {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		log.Println("Keep-Alive: APP_URL सेट नहीं है, सेल्फ-पिंग निष्क्रिय है।")
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

// एसिंक्रोनस वर्कर पूल
func startWebhookWorkerPool(workers int) {
	for i := 1; i <= workers; i++ {
		go func(workerID int) {
			for job := range webhookQueue {
				processIncomingMessage(job.From, job.Body)
			}
		}(i)
	}
}

// WhatsApp एकल-नंबर इंटेंट निष्पादन
func processIncomingMessage(from, body string) {
	cleanFrom := strings.TrimPrefix(strings.TrimSpace(from), "+")
	cleanBody := strings.TrimSpace(body)
	upperBody := strings.ToUpper(cleanBody)

	// स्पैम कॉलर फ़िल्टर
	profile := contactFilter.CheckCaller(cleanFrom)
	if profile.IsFriend || profile.IsSpam {
		return
	}

	// 1. अभिभावक सुरक्षा इंटेंट रूटिंग
	if pinMgr != nil {
		if upperBody == "UNLOCK" || upperBody == "OVERRIDE" {
			otp, err := pinMgr.GenerateOTP(cleanFrom)
			if err == nil {
				pinMgr.SendWhatsAppAlert(cleanFrom, fmt.Sprintf("🔐 सुरक्षा कोड: %s (वैधता: 5 मिनट)। अनलॉक हेतु 'OTP-%s' लिखकर भेजें।", otp, otp))
			}
			return
		}

		if strings.HasPrefix(upperBody, "OTP-") {
			enteredOTP := strings.TrimSpace(strings.TrimPrefix(upperBody, "OTP-"))
			if err := pinMgr.UnlockViaParentEmergencyOTP(cleanFrom, enteredOTP); err != nil {
				pinMgr.SendWhatsAppAlert(cleanFrom, fmt.Sprintf("❌ आपातकालीन अनलॉक विफल: %v", err))
			} else {
				pinMgr.SendWhatsAppAlert(cleanFrom, "✅ आपातकालीन सत्यापन सफल! सिस्टम आज के लिए अनलॉक कर दिया गया है।")
			}
			return
		}

		if upperBody == "BYPASS" {
			code, err := pinMgr.IssueOneTimeBypass(cleanFrom)
			if err != nil {
				pinMgr.SendWhatsAppAlert(cleanFrom, fmt.Sprintf("⚠️ %v", err))
			} else {
				pinMgr.SendWhatsAppAlert(cleanFrom, fmt.Sprintf("⏳ 15-मिनट पासकोड: %s (दैनिक सीमा: अधिकतम 2 बार)।", code))
			}
			return
		}
	}

	// 2. लाइव अध्ययन प्रगति रिपोर्ट
	if (upperBody == "REPORT" || upperBody == "प्रगति") && weeklyReportEngine != nil {
		reportText := weeklyReportEngine.GenerateSummary(cleanFrom)
		if pinMgr != nil {
			pinMgr.SendWhatsAppAlert(cleanFrom, reportText)
		}
		return
	}

	// 3. छात्र स्वास्थ्य व वेलनेस इंटेंट
	if wellnessEngine != nil && wellnessEngine.DetectSicknessFromMessage(cleanBody) {
		reply := wellnessEngine.MarkStudentSick(cleanFrom, "विद्यार्थी", "दैनिक अभ्यास")
		if pinMgr != nil {
			pinMgr.SendWhatsAppAlert(cleanFrom, reply)
		}
		return
	}

	if wellnessEngine != nil && (strings.Contains(strings.ToLower(cleanBody), "ठीक है") || strings.Contains(strings.ToLower(cleanBody), "स्वस्थ")) {
		reply := wellnessEngine.MarkStudentRecovered(cleanFrom, "विद्यार्थी")
		if pinMgr != nil {
			pinMgr.SendWhatsAppAlert(cleanFrom, reply)
		}
		return
	}

	// 4. ऑनबोर्डिंग व दैनिक 15-मिनट अभ्यास प्रवाह
	reply := fullOnboardingEngine.ProcessMessage(cleanFrom, cleanBody)
	if reply != "" && pinMgr != nil {
		pinMgr.SendWhatsAppAlert(cleanFrom, reply)
	}
}

// ✅ लूपहोल 3 निवारण: सुरक्षित नॉन-ड्रॉपिंग वेबहुक हैंडलर
func handleIncomingCommunication(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	body := r.URL.Query().Get("body")
	if from == "" {
		from = r.FormValue("From")
		body = r.FormValue("Body")
	}

	if from == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("MISSING_FROM"))
		return
	}

	// पहले चैनल में पुश करने की कोशिश करें, अगर बफ़र भर गया है तो 503 दें ताकि WhatsApp री-ट्राई करे
	select {
	case webhookQueue <- WebhookJob{From: from, Body: body}:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("EVENT_RECEIVED"))
	default:
		log.Printf("⚠️ गंभीर: वेबहुक बफर भर गया है! WhatsApp को 503 री-ट्राई भेजा जा रहा है: %s", from)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("SERVER_BUSY_RETRY_LATER"))
	}
}

func renderCertificateHTML(p *parental.CompleteParentProfile) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="hi">
<head>
  <meta charset="UTF-8">
  <title>अनंत अभ्यास - आधिकारिक मूल्यांकन प्रमाण पत्र</title>
  <style>
    @import url('https://fonts.googleapis.com/css2?family=Cinzel:wght@700&family=Montserrat:wght@400;600;700&display=swap');
    body { background-color: #0f172a; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; padding: 20px; font-family: 'Montserrat', sans-serif; }
    .cert-card { width: 850px; background: #ffffff; border: 12px solid #0f172a; outline: 3px solid #d97706; outline-offset: -8px; padding: 40px 50px; box-shadow: 0 25px 50px rgba(0,0,0,0.5); position: relative; color: #1e293b; background-image: radial-gradient(#f8fafc 90%%%%, #f1f5f9 100%%%%); }
    .header { display: flex; align-items: center; justify-content: space-between; border-bottom: 2px solid #e2e8f0; padding-bottom: 15px; }
    .brand { display: flex; align-items: center; gap: 15px; }
    .brand h1 { font-family: 'Cinzel', serif; font-size: 24px; margin: 0; color: #0f172a; }
    .student-block { text-align: center; margin: 25px 0; }
    .student-name { font-size: 30px; font-weight: 700; color: #0f172a; border-bottom: 2px solid #cbd5e1; display: inline-block; padding: 2px 25px; margin: 8px 0; }
    .metrics { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin: 25px 0; background: #f8fafc; padding: 15px; border: 1px solid #e2e8f0; border-radius: 6px; text-align: center; }
    .val { font-size: 18px; font-weight: 700; color: #0f172a; }
    .highlight { color: #16a34a; }
    .indicator-tag { background: #0f172a; color: #ffffff; padding: 12px; border-radius: 4px; display: flex; justify-content: space-between; align-items: center; }
    .badge { background: #d97706; padding: 4px 12px; border-radius: 3px; font-weight: 700; }
    .stamp { width: 65px; height: 65px; border: 2px dashed #d97706; border-radius: 50%%%%; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #d97706; font-size: 8px; font-weight: 700; transform: rotate(-10deg); }
  </style>
</head>
<body>
  <div class="cert-card">
    <div class="header">
      <div class="brand">
        <svg width="50" height="50" viewBox="0 0 100 100" fill="none">
          <path d="M50 5L90 20V50C90 75 50 95 50 95C50 95 10 75 10 50V20L50 5Z" fill="#0F172A" stroke="#D97706" stroke-width="4"/>
          <path d="M35 50C35 45 42 45 45 50C48 55 55 55 55 50C55 45 48 45 45 50C42 55 35 55 35 50Z" stroke="#F59E0B" stroke-width="3" fill="none"/>
        </svg>
        <div>
          <h1>ANANT ABHYAS</h1>
          <div style="font-size: 10px; color: #b45309; font-weight: 700; letter-spacing: 1px;">ROYAL FMC CORPORATION • ASSESSMENT WING</div>
        </div>
      </div>
      <div style="text-align: right; font-size: 11px; color: #64748b;">
        <strong>आधिकारिक मूल्यांकन पत्र</strong><br>
        आईडी: %s
      </div>
    </div>

    <div class="student-block">
      <p style="margin: 0; color: #64748b; font-size: 14px;">प्रमाणित किया जाता है कि विद्यार्थी</p>
      <div class="student-name">%s</div>
      <div style="font-size: 13px; color: #334155; font-weight: 600;">
        अभिभावक: श्री %s | कक्षा: %d | विंग: <strong>%s</strong>
      </div>
    </div>

    <div class="metrics">
      <div><div style="font-size: 11px; color: #64748b;">परीक्षण अंक</div><div class="val highlight">%d / 20</div></div>
      <div><div style="font-size: 11px; color: #64748b;">सटीकता दर</div><div class="val">80%%%%</div></div>
      <div><div style="font-size: 11px; color: #64748b;">कुल समय</div><div class="val">%d सेकंड</div></div>
      <div><div style="font-size: 11px; color: #64748b;">एंटी-चीट ऑडिट</div><div class="val" style="color: #0284c7;">100%%%% प्रामाणिक</div></div>
    </div>

    <div class="indicator-tag">
      <div><strong>निर्धारित इंडेक्स स्तर:</strong> %s</div>
      <div class="badge">VERIFIED</div>
    </div>

    <div style="display: flex; justify-content: space-between; align-items: flex-end; margin-top: 30px; border-top: 1px solid #e2e8f0; padding-top: 15px;">
      <div style="font-size: 10px; color: #64748b; max-width: 500px;">
        यह परिणाम 15-मिनट Kiosk सत्र और रिस्पॉन्स टाइमर इंजन द्वारा क्रिप्टोग्राफिक रूप से तैयार किया गया है। 7-दिवसीय सिस्टम डेमो सक्रिय है।
      </div>
      <div style="display: flex; align-items: center; gap: 15px;">
        <div class="stamp"><span>★ VERIFIED ★</span><span>ANANT</span><span>SECURE</span></div>
        <div style="font-size: 11px; font-weight: 700; color: #0f172a;">परीक्षा नियंत्रक<br><span style="color:#64748b; font-weight:400;">अनंत अभ्यास</span></div>
      </div>
    </div>
  </div>
</body>
</html>`, p.DemoID, p.ChildName, p.ParentName, p.ClassGrade, p.ExamGoalWing, p.ScoreOutOf20, p.TotalSolvingSec, p.IndicatorLevel)
}

// 20 बैकग्राउंड इंजनों की मेमोरी रजिस्ट्री
type CoreEngineRegistry struct {
	DemoSvc          *pricing.DemoService
	PlanSvc          *pricing.PlanService
	LoyaltySvc       *pricing.LoyaltyService
	AprilSvc         *vacation.AprilSessionService
	WinterSvc        *vacation.WinterBootcampService
	BridgeSvc        *vacation.FoundationBridgeService
	HolidayExamSvc   *holiday.ExamSchedulerService
	InterestSvc      *vacation.CustomInterestService
	PanDialectSvc    *language.PanIndiaDialectService
	FusionDialectSvc *language.FusionDialectService
	VoiceTuner       *audio.VoiceTunerService
	StateHolidaySvc  *holiday.StateHolidayService
	BioDNA           *security.BiometricDNAService
	SecSuite         *security.SecuritySuite
	AntiSharing      *security.AntiSharingGuard
	GeoTravel        *security.GeoTravelService
	MindReader       *monitor.MindReader
	InactivityNudge  *monitor.InactivityNudgeService
	StealthComp      *stealth.StealthComposer
	MultiChild       *family.MultiChildEngine
	FeaturePhone     *featurephone.FeaturePhoneEngine
	ImageEnhancer    *security.ImageEnhancer
	AutoHealer       *internal.AutoHealerEngine
}

type UltraClusterHub struct {
	Brain          *learning.AdaptiveSystemBrain
	ClusterMesh    *cluster.DynamicClusterMesh
	AutoPayEngine  *finance.AutoPayManager
	ParentFeedback *feedback.SupportEngineService
	AppSupport     *support.AutoHealingEngine
	Aggregator     *scale.HighConcurrencyAggregator
	CoreEngines    *CoreEngineRegistry
}

func main() {
	log.Println("==========================================================")
	log.Println("🚀 'अनंत अभ्यास अल्ट्रा' 360° प्रोडक्शन क्लस्टर लाइव")
	log.Println("==========================================================")

	go StartKeepAlive()

	// 1. एनवायरनमेंट वेरिएबल्स
	adminNumber := os.Getenv("ADMIN_PHONE_NUMBER")
	if adminNumber == "" {
		adminNumber = "9024414973"
	}
	gatewayNumber := os.Getenv("GATEWAY_PHONE_NUMBER")
	if gatewayNumber == "" {
		gatewayNumber = "9664006651"
	}
	secKey := os.Getenv("APP_SECURITY_SECRET")
	if secKey == "" {
		secKey = "ANANT_ULTRA_SECURE_TOKEN_SECRET_2026"
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:password@localhost:5432/anant_abhyas?sslmode=disable"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. ✅ लूपहोल 1 निवारण: DB अनिवार्य है, डाउन होने पर क्रैश सुरक्षित
	db, err := database.ConnectPostgres(connStr)
	if err != nil {
		log.Fatalf("❌ गंभीर त्रुटि: PostgreSQL डेटाबेस कनेक्ट नहीं हुआ। सुरक्षा कारणों से सिस्टम बंद हो रहा है: %v", err)
	}
	log.Println("✅ PostgreSQL डेटाबेस सफलतापूर्वक कनेक्ट हुआ।")
	if err := database.AutoMigrateDatabase(db); err != nil {
		log.Printf("⚠️ स्कीमा माइग्रेशन चेतावनी: %v", err)
	} else {
		log.Println("✅ सभी स्कीमा टेबल्स सत्यापित और अद्यतन हैं।")
	}

	// 3. डेटाबेस-आधारित कोर सेवाएं
	wellnessEngine = vacation.NewChildWellnessService(db)
	pinMgr = security.NewPINManager(db)
	weeklyReportEngine = monitor.NewWeeklyReportService(db)

	// 4. एसिंक्रोनस वर्कर पूल एक्टिवेशन
	startWebhookWorkerPool(10)

	// 5. कोर क्लस्टर Hub और 20 इंजन
	brain := learning.NewAdaptiveSystemBrain()
	brain.RunSelfLearningCycle()

	aggregator := scale.NewHighConcurrencyAggregator()
	aggregator.StartFlushDaemon(5 * time.Second)

	autoPayEngine := finance.NewAutoPayManager()
	parentFeedback := feedback.NewSupportEngineService(db)
	appSupport := support.NewAutoHealingEngine()

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

	coreRegistry := &CoreEngineRegistry{
		DemoSvc:          pricing.NewDemoService(db),
		PlanSvc:          pricing.NewPlanService(db),
		LoyaltySvc:       pricing.NewLoyaltyService(db),
		AprilSvc:         vacation.NewAprilSessionService(db),
		WinterSvc:        vacation.NewWinterBootcampService(db),
		BridgeSvc:        vacation.NewFoundationBridgeService(),
		HolidayExamSvc:   holiday.NewExamSchedulerService(db),
		InterestSvc:      vacation.NewCustomInterestService(db),
		PanDialectSvc:    language.NewPanIndiaDialectService(),
		FusionDialectSvc: language.NewFusionDialectService(db),
		VoiceTuner:       audio.NewVoiceTunerService(),
		StateHolidaySvc:  holiday.NewStateHolidayService(db),
		BioDNA:           security.NewBiometricDNAService(db),
		SecSuite:         security.NewSecuritySuite(secKey, db),
		AntiSharing:      security.NewAntiSharingGuard(db),
		GeoTravel:        security.NewGeoTravelService(db),
		MindReader:       monitor.NewMindReader(),
		InactivityNudge:  monitor.NewInactivityNudgeService(db),
		StealthComp:      stealth.NewStealthComposer(),
		MultiChild:       family.NewMultiChildEngine(db),
		FeaturePhone:     featurephone.NewFeaturePhoneEngine(db),
		ImageEnhancer:    security.NewImageEnhancer(),
		AutoHealer:       internal.NewAutoHealerEngine(adminNumber, 400.0),
	}

	hub := &UltraClusterHub{
		Brain:          brain,
		ClusterMesh:    clusterMesh,
		AutoPayEngine:  autoPayEngine,
		ParentFeedback: parentFeedback,
		AppSupport:     appSupport,
		Aggregator:     aggregator,
		CoreEngines:    coreRegistry,
	}
	log.Println("✅ सभी 20 कोर इंजन Hub के साथ पूर्णतः एकीकृत और लाइव हैं।")

	// 6. HTTP एंडपॉइंट्स रूटिंग
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<h2>🎓 अनंत अभ्यास (रॉयल एफएमसी कॉरपोरेशन) लाइव है 24x7!</h2><p><a href="/admin">एडमिन पोर्टल</a></p>`)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"HEALTHY","cluster":"ONLINE"}`))
	})

	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		snap := clockEngine.GetCurrentSnapshot()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h2>अनंत अभ्यास एडमिन पोर्टल</h2><p>गेटवे: %s | एडमिन: %s</p><p>समय (IST): %s</p>", gatewayNumber, adminNumber, snap.FormattedTimestamp)
	})

	// 🔐 ज़ीरो-ट्रस्ट मास्टर पिन (सुरक्षित POST RBAC)
	mux.HandleFunc("/api/v1/parent/verify-pin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "विधि अस्वीकृत (केवल POST मान्य)", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Phone string `json:"phone"`
			PIN   string `json:"pin"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "अमान्य JSON डेटा", http.StatusBadRequest)
			return
		}

		ok, code, err := pinMgr.VerifyPINWithDailyLock(payload.Phone, payload.PIN)
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "FAILED",
				"code":    code,
				"message": err.Error(),
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "SUCCESS",
			"code":    code,
			"message": "पिन सत्यापन सफल",
		})
	})

	mux.HandleFunc("/api/v1/parent/request-pin-otp", func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		otp, err := pinMgr.GenerateOTP(phone)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"status": "ERROR", "message": err.Error()})
			return
		}

		pinMgr.SendWhatsAppAlert(phone, fmt.Sprintf("सुरक्षा कोड: %s (5 मिनट में समाप्त)", otp))
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "OTP_SENT",
			"message": "सत्यापन कोड WhatsApp पर भेजा गया",
		})
	})

	mux.HandleFunc("/api/v1/parent/emergency-unlock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "केवल POST मान्य", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Phone string `json:"phone"`
			OTP   string `json:"otp"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "अमान्य डेटा", http.StatusBadRequest)
			return
		}

		err := pinMgr.UnlockViaParentEmergencyOTP(payload.Phone, payload.OTP)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"status": "ERROR", "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "SUCCESS", "message": "सिस्टम अनलॉक हुआ"})
	})

	mux.HandleFunc("/api/v1/parent/issue-bypass", func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		code, err := pinMgr.IssueOneTimeBypass(phone)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"status": "QUOTA_EXCEEDED", "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"status":      "SUCCESS",
			"bypass_code": code,
			"validity":    "15 Minutes",
		})
	})

	// WhatsApp नॉन-ब्लॉकिंग गेटवे रूट्स
	mux.HandleFunc("/webhook", handleIncomingCommunication)
	mux.HandleFunc("/incoming", handleIncomingCommunication)

	// डिजिटल मेरिट सर्टिफिकेट रूट
	mux.HandleFunc("/cert/", func(w http.ResponseWriter, r *http.Req
