package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"anant-project/audio"
	"anant-project/database"
	"anant-project/family"
	"anant-project/featurephone"
	"anant-project/feedback"
	"anant-project/holiday"
	"anant-project/language"
	"anant-project/monitor"
	"anant-project/pricing"
	"anant-project/security"
	"anant-project/stealth"
	"anant-project/temporal"
	"anant-project/vacation"
)

func main() {
	fmt.Println("🚀 'अनंत अभ्यास' 360° मास्टर प्रोडक्शन बैकएंड प्रारंभ हो रहा है...")

	// 1. डेटाबेस इनिशियलाइज़ेशन (PostgreSQL / Fallback)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:password@localhost:5432/anant_abhyas?sslmode=disable"
	}
	db := database.InitDB(connStr)
	defer db.Close()

	// 2. केंद्रीय समय व कैलेंडर इंटेलिजेंस इंजन (Asia/Kolkata Lock)
	clockEngine := temporal.NewClockEngine()
	timeSnap := clockEngine.GetCurrentSnapshot()

	// 3. सभी मॉड्यूल्स व सेवाओं का लिंकेज
	demoService := pricing.NewDemoService(db)
	planService := pricing.NewPlanService(db)
	loyaltyService := pricing.NewLoyaltyService(db)
	aprilSession := vacation.NewAprilSessionService(db)
	homeworkDecomposer := vacation.NewHomeworkService(db)
	winterBootcamp := vacation.NewWinterBootcampService(db)
	foundationBridge := vacation.NewFoundationBridgeService()
	examScheduler := holiday.NewExamSchedulerService(db)
	customInterest := vacation.NewCustomInterestService(db)
	pacingService := vacation.NewPacingService()
	panIndiaDialect := language.NewPanIndiaDialectService()
	fusionEngine := language.NewFusionDialectService(db)
	voiceTuner := audio.NewVoiceTunerService()
	supportEngine := feedback.NewSupportEngineService(db)
	stateEngine := holiday.NewStateHolidayService(db)
	biometricDNA := security.NewBiometricDNAService(db)
	secSuite := security.NewSecuritySuite("ANANT_MASTER_SECRET_2026", db)
	antiSharing := security.NewAntiSharingGuard(db)
	geoGuard := security.NewGeoTravelService(db)
	mindReader := monitor.NewMindReader()
	inactivityNudge := monitor.NewInactivityNudgeService()
	stealthComposer := stealth.NewStealthComposer()

	// 4. नए एक्सटेंडेड मॉड्यूल्स (Family Multi-Child, IVR/SMS, Image Enhancer, Weekly Report)
	multiChildService := family.NewMultiChildEngine(db)
	ivrService := featurephone.NewFeaturePhoneEngine(db)
	imageEnhancer := security.NewImageEnhancer()
	weeklyReportService := monitor.NewWeeklyReportService(db)

	// अप्रयुक्त वेरिएबल वार्निंग से बचने के लिए
	_ = demoService
	_ = homeworkDecomposer
	_ = foundationBridge
	_ = pacingService
	_ = panIndiaDialect
	_ = mindReader
	_ = inactivityNudge

	// 5. टेस्ट सबमिशन सिमुलेशन (7-डे डेमो + 1 फोन 4 बच्चे + लाइव टाइम + बाल-स्वरुचि)
	studentPhone := "9024414973"
	studentName := "आरव"
	registeredGrade := 6
	scannedPageGrade := 6
	studentState := "Rajasthan"
	studentDistrict := "Sri Ganganagar"
	currentTier := pricing.TierDemo
	currentDeviceHash := "DEVICE_SAMSUNG_M31_ORIGINAL"
	rawInputSpeech := "मैंने ठा कोनी दस मैं की कहना चाहना"
	childChosenTopicVoice := "मुझे कारों के इंजन और रोबोटिक्स के बारे में सीखना है"
	sampleOCR := "प्रश्न 1: भिन्न का जोड़। 1/2 + 3/4 = 5/4।"
	dummyImageBytes := []byte("image_payload_stream_bytes_sample_valid_length_data")
	handwritingSimilarity := 0.94

	// A. इमेज क्वालिटी हीलर व प्री-प्रोसेसर
	imgValid, imgQualityMsg := imageEnhancer.CheckAndNormalizeQuality(dummyImageBytes)
	if !imgValid {
		fmt.Printf("⚠️ इमेज क्वालिटी अलर्ट: %s\n", imgQualityMsg)
	}

	// B. सुरक्षा, रेट लिमिट व एंटी-शेयरिंग
	limits := planService.GetPlanLimits(currentTier)
	allowed, limitMsg := secSuite.ValidateRateLimit(studentPhone, limits.MaxDailyScans)
	if !allowed {
		fmt.Printf("🛑 सुरक्षा ब्लॉक: %s\n", limitMsg)
	}
	sharingVerdict := antiSharing.EvaluateSharingRisk(studentPhone, registeredGrade, scannedPageGrade, handwritingSimilarity, 1, limits.MaxDailyScans)
	if sharingVerdict.IsBlocked {
		fmt.Printf("🛑 शेयरिंग ब्लॉक: %s\n", sharingVerdict.UserMessage)
	}
	_, _ = biometricDNA.VerifyHandwritingDNA(studentPhone, handwritingSimilarity)

	// C. हैश व डुप्लिकेट सत्यापन
	imgHash := secSuite.GenerateHash(dummyImageBytes)
	_, _, userMsg := secSuite.VerifySubmission(studentPhone, sampleOCR, imgHash)

	// D. बाल-स्वरुचि ट्रैक व राज्य-वार सत्र
	childTrack := customInterest.AutoSetFromChildVoice(studentPhone, childChosenTopicVoice)
	sessionInfo := aprilSession.ResolveStateAcademicSession(studentState, studentDistrict)
	winterInfo := winterBootcamp.ResolveWinterBootcamp(studentState, studentDistrict)

	// E. 1 फ़ोन 4 बच्चे (1-Hour Time-Sliced Multi-Child Engine)
	children := []family.ChildSession{
		{ChildID: "C1", Name: "आरव", Grade: 6},
		{ChildID: "C2", Name: "दीया", Grade: 4},
		{ChildID: "C3", Name: "कबीर", Grade: 8},
		{ChildID: "C4", Name: "अनन्या", Grade: 2},
	}
	schedule := multiChildService.GenerateOneHourSchedule(studentPhone, children)
	_, slotStatus := multiChildService.GetCurrentActiveChild(schedule)

	// F. परीक्षा, बोली, अवकाश व लॉयल्टी विश्लेषण
	examStatus := examScheduler.GetActiveExamMode(studentPhone, rawInputSpeech)
	dialectProf := fusionEngine.DetectAndResolve(rawInputSpeech, studentState, studentDistrict)
	isHoliday, holidayType, daysLeft := stateEngine.CheckHoliday(studentState, studentDistrict)
	renewalFee, loyaltyMsg := loyaltyService.CalculateLoyaltyFee(studentPhone, 149.0, currentTier)

	// G. माता-पिता फीडबैक व जिओ-ट्रैवल
	parentFeedback := "मास्टरजी, शाम ने 6 बजे कटाई रो काम चाले है, टेम 8 बजे रो कर सको के?"
	ticket := supportEngine.ProcessFeedback(studentPhone, studentState, studentDistrict, dialectProf.DialectCode, parentFeedback)
	_ = geoGuard.CheckTravelEvent(studentPhone, studentDistrict, currentDeviceHash, handwritingSimilarity)

	// H. स्टील्थ AI प्रॉम्प्ट व वॉइस SSML
	kidSSML := voiceTuner.GenerateKidFriendlySSML(studentName, "आज आपने भिन्न का जोड़ सही लिखा है।", dialectProf.ToneHint)
	aiPrompt := stealthComposer.BuildPrompt(studentName, studentState, studentDistrict, sessionInfo.SessionPhase, string(currentTier), isHoliday, dialectProf)

	// I. रविवार साप्ताहिक रिपोर्ट कार्ड & फ़ीचर फ़ोन IVR स्क्रिप्ट
	weeklyProgress := weeklyReportService.GenerateSundayReport(studentPhone, studentName)
	whatsAppReport := weeklyReportService.FormatWhatsAppReportCard(weeklyProgress)
	ivrScript := ivrService.GenerateIVRScript(studentName, dialectProf.DialectCode, "1/2 में 3/4 जोड़ने पर क्या आएगा?")

	// 6. संपूर्ण मास्टर सिस्टम रिपोर्ट (कंसोल लॉग)
	fmt.Println("\n=======================================================")
	fmt.Printf("⏱️ लाइव सर्वर टाइम: %s (%s)\n", timeSnap.FormattedTimestamp, timeSnap.AcademicSessionLabel)
	fmt.Printf("📅 कल की तारीख: %s | माह अंत: %t | वर्ष अंत: %t\n", timeSnap.NextDayDateString, timeSnap.IsMonthLastDay, timeSnap.IsYearLastDay)
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("✅ सबमिशन स्थिति: %s (शेयरिंग रिस्क: 0%%)\n", userMsg)
	fmt.Printf("👤 प्लान: %s (वैधता: %d दिन, कोटा: %d स्कैन/दिन)\n", currentTier, limits.TrialDays, limits.MaxDailyScans)
	fmt.Printf("👨‍👩‍👧‍👦 1-घंटा 4-बच्चे स्लॉट: %s\n", slotStatus)
	fmt.Printf("📍 राज्य-वार सत्र: %s\n", sessionInfo.SessionPhase)
	fmt.Printf("❄️ राज्य-वार विंटर स्थिति: %t (शेष दिन: %d)\n", winterInfo.IsWinterBreak, winterInfo.DaysLeft)
	fmt.Printf("🎯 परीक्षा मोड: %s\n", examStatus.ExamHeadline)
	fmt.Printf("🌟 बाल-स्वरुचि विषय: \"%s\"\n", childTrack.TopicName)
	fmt.Printf("🗣️ सक्रिय बोली: %s (%s)\n", dialectProf.DialectCode, dialectProf.RegionHint)
	fmt.Printf("📅 अवकाश स्थिति: %t (%s, शेष दिन: %d)\n", isHoliday, holidayType, daysLeft)
	fmt.Printf("💰 नवीनीकरण स्थिति: ₹%.2f (%s)\n", renewalFee, loyaltyMsg)
	fmt.Printf("📩 सपोर्ट टिकट: %s (गंभीरता: %d/5)\n", ticket.Category, ticket.UrgencyScore)
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("🎙️ वॉइस SSML:\n%s\n", kidSSML)
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("☎️ फ़ीचर फ़ोन IVR डायलॉग:\n\"%s\"\n", ivrScript)
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("📊 WhatsApp साप्ताहिक रिपोर्ट कार्ड:\n%s\n", whatsAppReport)
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("🤖 मास्टर स्टील्थ AI प्रॉम्प्ट:\n%s\n", aiPrompt)
	fmt.Println("=======================================================")

	// ==================== 7. लाइव HTTP सर्वर राउट्स (Render & Web Portal) ====================

	// 1. होम रूट (Render Health Check)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<h2>🎓 अनंत अभ्यास (Anant Abhyas) सर्वर 100% सक्रिय है।</h2><p>एडमिन कंट्रोल पोर्टल के लिए <a href="/admin">यहाँ क्लिक करें (/admin)</a></p>`)
	})

	// 2. vCard 1-टैप कॉन्टैक्ट सेवर
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		vcard := strings.Join([]string{
			"BEGIN:VCARD",
			"VERSION:3.0",
			"N:मास्टरजी;अनंत अभ्यास;;;",
			"FN:अनंत अभ्यास - डिजिटल मास्टरजी",
			"ORG:Anant Abhyas Education;",
			"TEL;TYPE=CELL,VOICE,PREF:+919664006651",
			"NOTE:रोजाना 15 मिनट बोलकर अभ्यास और 7-दिन फ्री डेमो।",
			"URL:https://wa.me/919664006651?text=राम%20राम%20सा%20मुझे%20फ्री%20डेमो%20चाहिए",
			"END:VCARD",
		}, "\r\n")

		w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\"Anant_Abhyas.vcf\"")
		w.Write([]byte(vcard))
	})

	// 3. एडमिन वेब पोर्टल (/admin)
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html>
<html lang="hi">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>अनंत अभ्यास एडमिन पोर्टल</title>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; background: #0f172a; color: #f8fafc; padding: 16px; margin: 0; }
        .card { background: #1e293b; padding: 16px; border-radius: 12px; margin-bottom: 16px; }
        .btn-wa { display: block; background: #25d366; color: white; text-align: center; padding: 14px; border-radius: 8px; font-weight: bold; text-decoration: none; margin-bottom: 16px; }
        .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
        .metric { background: #0f172a; padding: 12px; border-radius: 8px; text-align: center; }
        .val { font-size: 24px; font-weight: bold; color: #38bdf8; margin-top: 4px; }
    </style>
</head>
<body>
    <h2>🎓 अनंत अभ्यास एडमिन</h2>
    <p style="color: #4ade80; margin-top: -10px;">गेटवे: +91 9664006651 (Jio Live)</p>

    <a class="btn-wa" href="https://wa.me/919664006651?text=राम%20राम%20सा%2C%20मुझे%20अनंत%20अभ्यास%20का%20फ्री%20डेमो%20चाहिए" target="_blank">
        📲 WhatsApp चैट लिंक खोलें
    </a>

    <div class="card">
        <h3>📊 लाइव मेट्रिक्स</h3>
        <div class="grid">
            <div class="metric"><small>कुल सक्रिय छात्र</small><div class="val">1</div></div>
            <div class="metric"><small>नए जुड़े बच्चे</small><div class="val" style="color:#4ade80;">+1</div></div>
            <div class="metric"><small>रेफरल पूरे हुए</small><div class="val" style="color:#a855f7;">0</div></div>
            <div class="metric"><small>पेंडिंग रेफरल</small><div class="val" style="color:#facc15;">0</div></div>
        </div>
    </div>

    <div class="card">
        <h3>🔍 360° छात्र कुंडली सर्च</h3>
        <input type="text" id="q" placeholder="UID या फोन नंबर..." style="width: 90%; padding: 10px; background: #0f172a; border: 1px solid #475569; color: #fff; border-radius: 6px;">
        <button onclick="alert('छात्र: आरव (कक्षा 6) | स्थिति: एक्टिव डेमो (7 दिन)')" style="margin-top: 8px; width: 100%; padding: 10px; background: #2563eb; color: #fff; border: none; border-radius: 6px; font-weight: bold;">सर्च करें</button>
    </div>
</body>
</html>`)
	})

	// 8. पोर्ट बाइंडिंग व सर्वर लिसनर
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 अनंत अभ्यास लाइव वेब सर्वर पोर्ट :%s पर सक्रिय है...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("सर्वर त्रुटि: %v", err)
	}
}
