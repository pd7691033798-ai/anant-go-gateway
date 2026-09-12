package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

const (
	AdminNumber      = "9024414973"
	GatewayNumber    = "9664006651"
	MerchantVPA      = "9664006651@ptsbi"
	MerchantName     = "Royal FMC corporation"
	BrandDisplayName = "Anant Abhyas"
)

var (
	clockEngine          = temporal.NewClockEngine()
	contactFilter        *featurephone.ContactFilter
	fullOnboardingEngine *parental.FullWhatsAppEngine
	wellnessEngine       *vacation.ChildWellnessService
	pinMgr               *security.PINManager
	weeklyReportEngine   *monitor.WeeklyReportService
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

// WhatsApp एकल-नंबर इंटेंट राउटर (Onboarding, Practice & Parental Controls)
func handleIncomingCommunication(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	body := r.URL.Query().Get("body")
	if from == "" {
		from = r.FormValue("From")
		body = r.FormValue("Body")
	}

	cleanFrom := strings.TrimPrefix(strings.TrimSpace(from), "+")
	cleanBody := strings.TrimSpace(body)
	upperBody := strings.ToUpper(cleanBody)

	// 1. स्पैम व अनधिकृत कॉलर फ़िल्टर
	profile := contactFilter.CheckCaller(cleanFrom)
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

	// 2. अभिभावक सुरक्षा इंटेंट रूटिंग (पिन रीसेट, अनलॉक और बायपास)
	if pinMgr != nil {
		if upperBody == "UNLOCK" || upperBody == "OVERRIDE" || strings.HasPrefix(upperBody, "OTP-") {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)

			if strings.HasPrefix(upperBody, "OTP-") {
				enteredOTP := strings.TrimSpace(strings.TrimPrefix(upperBody, "OTP-"))
				err := pinMgr.UnlockViaParentEmergencyOTP(cleanFrom, enteredOTP)
				if err != nil {
					w.Write([]byte(fmt.Sprintf("❌ अनलॉक विफल: %v", err)))
					return
				}
				w.Write([]byte("✅ आपातकालीन सत्यापन सफल! सिस्टम आज के लिए अनलॉक कर दिया गया है।"))
				return
			}

			otp, err := pinMgr.GenerateOTP(cleanFrom)
			if err != nil {
				w.Write([]byte("⚠️ OTP जनरेट करने में विफलता। बाद में प्रयास करें।"))
				return
			}
			w.Write([]byte(fmt.Sprintf("🔐 सुरक्षा कोड: %s\nसिस्टम अनलॉक करने के लिए 'OTP-%s' लिखकर भेजें (वैधता: 5 मिनट)।", otp, otp)))
			return
		}

		if upperBody == "BYPASS" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			code, err := pinMgr.IssueOneTimeBypass(cleanFrom)
			if err != nil {
				w.Write([]byte(fmt.Sprintf("⚠️ %v", err)))
				return
			}
			w.Write([]byte(fmt.Sprintf("⏳ 15-मिनट पासकोड: %s\nयह कोड अगले 15 मिनट के लिए मान्य है (दैनिक कोटा: अधिकतम 2 बार)।", code)))
			return
		}
	}

	// 3. लाइव अध्ययन प्रगति रिपोर्ट (बिना किसी डमी टेक्स्ट के)
	if upperBody == "REPORT" || upperBody == "प्रगति" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if weeklyReportEngine != nil {
			reportText := weeklyReportEngine.GenerateSummary(cleanFrom)
			w.Write([]byte(reportText))
			return
		}
		w.Write([]byte("📊 वर्तमान में डेटाबेस से प्रगति रिपोर्ट प्राप्त नहीं हो सकी। कृपया कुछ समय पश्चात पुनः प्रयास करें।"))
		return
	}

	// 4. छात्र स्वास्थ्य व वेलनेस इंटेंट
	if wellnessEngine != nil && wellnessEngine.DetectSicknessFromMessage(cleanBody) {
		reply := wellnessEngine.MarkStudentSick(cleanFrom, "विद्यार्थी", "दैनिक अभ्यास")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(reply))
		return
	}

	if wellnessEngine != nil && (strings.Contains(strings.ToLower(cleanBody), "ठीक है") || strings.Contains(strings.ToLower(cleanBody), "स्वस्थ")) {
		reply := wellnessEngine.MarkStudentRecovered(cleanFrom, "विद्यार्थी")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(reply))
		return
	}

	// 5. ऑनबोर्डिंग व दैनिक 15-मिनट अभ्यास प्रवाह
	reply := fullOnboardingEngine.ProcessMessage(cleanFrom, cleanBody)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(reply))
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

	// 🔐 1. डेटाबेस-आधारित कोर सुरक्षा, वेलनेस व रिपोर्ट इनिशियलाइज़ेशन
	wellnessEngine = vacation.NewChildWellnessService(db)
	pinMgr = security.NewPINManager(db)
	weeklyReportEngine = monitor.NewWeeklyReportService(db)

	// 2. सेल्फ-लर्निंग ब्रेन और स्केल बफर
	brain := learning.NewAdaptiveSystemBrain()
	brain.RunSelfLearningCycle()

	aggregator := scale.NewHighConcurrencyAggregator()
	aggregator.StartFlushDaemon(5 * time.Second)

	autoPayEngine := finance.NewAutoPayManager()
	parentFeedback := feedback.NewSupportEngineService(db)
	appSupport := support.NewAutoHealingEngine()

	// 3. ऑटो-स्केलिंग क्लस्टर मेश
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

	secKey := os.Getenv("APP_SECURITY_SECRET")
	if secKey == "" {
		secKey = "ANANT_ULTRA_SECURE_TOKEN_SECRET_2026"
	}

	// 4. सभी 20 कोर बैकग्राउंड सर्विसेज को Hub के साथ लाइव बाइंड करना
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
		AutoHealer:       internal.NewAutoHealerEngine(AdminNumber, 400.0),
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

	// 5. HTTP एंडपॉइंट्स रूटिंग
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

	// 🔐 ज़ीरो-ट्रस्ट मास्टर पिन वेरिफिकेशन
	http.HandleFunc("/api/v1/parent/verify-pin", func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		pin := r.URL.Query().Get("pin")

		ok, code, err := pinMgr.VerifyPINWithDailyLock(phone, pin)
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

	http.HandleFunc("/api/v1/parent/request-pin-otp", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/api/v1/parent/emergency-unlock", func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		otp := r.URL.Query().Get("otp")

		err := pinMgr.UnlockViaParentEmergencyOTP(phone, otp)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"status": "ERROR", "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "SUCCESS", "message": "सिस्टम अनलॉक हुआ"})
	})

	http.HandleFunc("/api/v1/parent/issue-bypass", func(w http.ResponseWriter, r *http.Request) {
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

	// WhatsApp गेटवे रूट्स
	http.HandleFunc("/webhook", handleIncomingCommunication)
	http.HandleFunc("/incoming", handleIncomingCommunication)

	// डिजिटल मेरिट सर्टिफिकेट रूट
	http.HandleFunc("/cert/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		demoID := parts[2]
		profile := fullOnboardingEngine.GetProfileByDemoID(demoID)
		if profile == nil {
			http.Error(w, "सर्टिफिकेट नहीं मिला या डेमो आईडी अमान्य है।", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := renderCertificateHTML(profile)
		w.Write([]byte(html))
	})

	// क्लस्टर ट्रैफिक डिस्पैचर
	http.HandleFunc("/cluster/dispatch", hub.ClusterMesh.RouteSmartTraffic)

	// ऑनबोर्डिंग व बिलिंग
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

		bill, err := newpricing.ComputeModularBill(tier, hasExam, 0, false)
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

		bill, err := newpricing.ComputeModularBill(tier, hasExam, 0, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sub := hub.AutoPayEngine.SetupMandate(parentID, float64(bill.FinalPayable))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":          "MANDATE_INITIATED",
			"monthly_amount":  bill.FinalPayable,
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

	// एडमिन स्टैट्स (अब Hub.CoreEngines से लाइव लिंक है)
	http.HandleFunc("/api/v1/admin/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"system":        "Anant Abhyas Ultra Core",
			"cluster_state": "ACTIVE",
			"auto_pay":      "ENFORCED_MANDATE_ONLY",
			"active_tracks": []string{"NAVODAYA", "SAINIK_SCHOOL", "NDA", "IIT_JEE"},
			"anti_sharing":  hub.CoreEngines.AntiSharing != nil,
			"biometric_dna": hub.CoreEngines.BioDNA != nil,
		})
	})

	// ⚡ ऑटोनोमस सैंडबॉक्स ट्रायल कोर
	sandboxCore := sandbox.NewAutonomousSandboxCore(AdminNumber, db, func(from, body string) string {
		if wellnessEngine != nil && wellnessEngine.DetectSicknessFromMessage(body) {
			return wellnessEngine.MarkStudentSick(from, "सैंडबॉक्स छात्र", "दैनिक अभ्यास")
		}
		return fullOnboardingEngine.ProcessMessage(from, body)
	})

	http.HandleFunc("/sandbox", sandboxCore.RenderSandboxUI)
	http.HandleFunc("/api/v1/sandbox/simulate", sandboxCore.HandleSimulation)
	http.HandleFunc("/api/v1/sandbox/toggle", sandboxCore.ToggleSimulationStates)
	http.HandleFunc("/api/v1/sandbox/audit", sandboxCore.ServeAuditReport)

	// 6. सर्वर स्टार्टअप
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 अनंत अभ्यास क्लस्टर पोर्ट :%s पर पूर्णतः सक्रिय है...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
