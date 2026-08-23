package main

import (
	"fmt"

	"anant-project/audio"
	"anant-project/database"
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
	fmt.Println("ðŸš€ 'à¤…à¤¨à¤‚à¤¤ à¤…à¤­à¥à¤¯à¤¾à¤¸' 360Â° à¤®à¤¾à¤¸à¥à¤Ÿà¤° à¤ªà¥à¤°à¥‹à¤¡à¤•à¥à¤¶à¤¨ à¤¬à¥ˆà¤•à¤à¤‚à¤¡ à¤ªà¥à¤°à¤¾à¤°à¤‚à¤­ à¤¹à¥‹ à¤°à¤¹à¤¾ à¤¹à¥ˆ...")

	// 1. à¤¡à¥‡à¤Ÿà¤¾à¤¬à¥‡à¤¸ à¤‡à¤¨à¤¿à¤¶à¤¿à¤¯à¤²à¤¾à¤‡à¤œà¤¼à¥‡à¤¶à¤¨
	connStr := "postgres://postgres:password@localhost:5432/anant_abhyas?sslmode=disable"
	db := database.InitDB(connStr)
	defer db.Close()

	// 2. à¤•à¥‡à¤‚à¤¦à¥à¤°à¥€à¤¯ à¤¸à¤®à¤¯ à¤µ à¤•à¥ˆà¤²à¥‡à¤‚à¤¡à¤° à¤‡à¤‚à¤Ÿà¥‡à¤²à¤¿à¤œà¥‡à¤‚à¤¸ à¤‡à¤‚à¤œà¤¨
	clockEngine := temporal.NewClockEngine()
	timeSnap := clockEngine.GetCurrentSnapshot()

	// 3. à¤¸à¤­à¥€ 20 à¤¸à¥à¤µà¤¤à¤‚à¤¤à¥à¤° à¤®à¥‰à¤¡à¥à¤¯à¥‚à¤² à¤¸à¥‡à¤µà¤¾à¤“à¤‚ à¤•à¤¾ à¤²à¤¿à¤‚à¤•à¥‡à¤œ
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

	// 4. à¤Ÿà¥‡à¤¸à¥à¤Ÿ à¤¸à¤¬à¤®à¤¿à¤¶à¤¨ à¤¸à¤¿à¤®à¥à¤²à¥‡à¤¶à¤¨ (7-à¤¡à¥‡ à¤¡à¥‡à¤®à¥‹ + à¤²à¤¾à¤‡à¤µ à¤Ÿà¤¾à¤‡à¤® + à¤¬à¤¾à¤²-à¤¸à¥à¤µà¤°à¥à¤šà¤¿ + à¤Ÿà¥‡à¤¸à¥à¤Ÿ à¤®à¥‹à¤¡)
	studentPhone := "9024414973"
	studentName := "à¤†à¤°à¤µ"
	registeredGrade := 6
	scannedPageGrade := 6
	studentState := "Rajasthan"
	studentDistrict := "Sri Ganganagar"
	currentTier := pricing.TierDemo // 7-à¤¦à¤¿à¤¨ à¤¨à¤¿à¤ƒà¤¶à¥à¤²à¥à¤• à¤¡à¥‡à¤®à¥‹
	currentDeviceHash := "DEVICE_SAMSUNG_M31_ORIGINAL"
	rawInputSpeech := "à¤®à¥ˆà¤‚à¤¨à¥‡ à¤ à¤¾ à¤•à¥‹à¤¨à¥€ à¤¦à¤¸ à¤®à¥ˆà¤‚ à¤•à¥€ à¤•à¤¹à¤¨à¤¾ à¤šà¤¾à¤¹à¤¨à¤¾"
	childChosenTopicVoice := "à¤®à¥à¤à¥‡ à¤•à¤¾à¤°à¥‹à¤‚ à¤•à¥‡ à¤‡à¤‚à¤œà¤¨ à¤”à¤° à¤°à¥‹à¤¬à¥‹à¤Ÿà¤¿à¤•à¥à¤¸ à¤•à¥‡ à¤¬à¤¾à¤°à¥‡ à¤®à¥‡à¤‚ à¤¸à¥€à¤–à¤¨à¤¾ à¤¹à¥ˆ"
	sampleOCR := "à¤ªà¥à¤°à¤¶à¥à¤¨ 1: à¤­à¤¿à¤¨à¥à¤¨ à¤•à¤¾ à¤œà¥‹à¤¡à¤¼à¥¤ 1/2 + 3/4 = 5/4à¥¤"
	dummyImageBytes := []byte("image_payload_stream_bytes")
	handwritingSimilarity := 0.94

	// A. à¤¸à¥à¤°à¤•à¥à¤·à¤¾, à¤°à¥‡à¤Ÿ à¤²à¤¿à¤®à¤¿à¤Ÿ à¤µ à¤à¤‚à¤Ÿà¥€-à¤¶à¥‡à¤¯à¤°à¤¿à¤‚à¤—
	limits := planService.GetPlanLimits(currentTier)
	allowed, limitMsg := secSuite.ValidateRateLimit(studentPhone, limits.MaxDailyScans)
	if !allowed {
		fmt.Printf("ðŸ›‘ à¤¸à¥à¤°à¤•à¥à¤·à¤¾ à¤¬à¥à¤²à¥‰à¤•: %s\n", limitMsg)
		return
	}
	sharingVerdict := antiSharing.EvaluateSharingRisk(studentPhone, registeredGrade, scannedPageGrade, handwritingSimilarity, 1, limits.MaxDailyScans)
	if sharingVerdict.IsBlocked {
		fmt.Printf("ðŸ›‘ à¤¶à¥‡à¤¯à¤°à¤¿à¤‚à¤— à¤¬à¥à¤²à¥‰à¤•: %s\n", sharingVerdict.UserMessage)
		return
	}
	_, _ = biometricDNA.VerifyHandwritingDNA(studentPhone, handwritingSimilarity)

	// B. à¤¹à¥ˆà¤¶ à¤µ à¤¡à¥à¤ªà¥à¤²à¤¿à¤•à¥‡à¤Ÿ à¤¸à¤¤à¥à¤¯à¤¾à¤ªà¤¨
	imgHash := secSuite.GenerateHash(dummyImageBytes)
	isValid, _, userMsg := secSuite.VerifySubmission(studentPhone, sampleOCR, imgHash)
	if !isValid {
		fmt.Println("âš ï¸ à¤¸à¤¬à¤®à¤¿à¤¶à¤¨ à¤…à¤¸à¥à¤µà¥€à¤•à¥ƒà¤¤")
		return
	}

	// C. à¤¬à¤¾à¤²-à¤¸à¥à¤µà¤°à¥à¤šà¤¿ à¤Ÿà¥à¤°à¥ˆà¤• à¤µ à¤°à¤¾à¤œà¥à¤¯-à¤µà¤¾à¤° à¤¸à¤¤à¥à¤°
	childTrack := customInterest.AutoSetFromChildVoice(studentPhone, childChosenTopicVoice)
	sessionInfo := aprilSession.ResolveStateAcademicSession(studentState, studentDistrict)
	winterInfo := winterBootcamp.ResolveWinterBootcamp(studentState, studentDistrict)

	// D. à¤ªà¤°à¥€à¤•à¥à¤·à¤¾, à¤¬à¥‹à¤²à¥€ à¤µ à¤…à¤µà¤•à¤¾à¤¶ à¤µà¤¿à¤¶à¥à¤²à¥‡à¤·à¤£
	examStatus := examScheduler.GetActiveExamMode(studentPhone, rawInputSpeech)
	dialectProf := fusionEngine.DetectAndResolve(rawInputSpeech, studentState, studentDistrict)
	isHoliday, holidayType, daysLeft := stateEngine.CheckHoliday(studentState, studentDistrict)
	renewalFee, loyaltyMsg := loyaltyService.CalculateLoyaltyFee(studentPhone, 149.0, currentTier)

	// E. à¤¸à¥à¤Ÿà¥€à¤²à¥à¤¥ AI à¤ªà¥à¤°à¥‰à¤®à¥à¤ªà¥à¤Ÿ à¤µ à¤µà¥‰à¤‡à¤¸ SSML
	kidSSML := voiceTuner.GenerateKidFriendlySSML(studentName, "à¤†à¤œ à¤†à¤ªà¤¨à¥‡ à¤­à¤¿à¤¨à¥à¤¨ à¤•à¤¾ à¤œà¥‹à¤¡à¤¼ à¤¸à¤¹à¥€ à¤²à¤¿à¤–à¤¾ à¤¹à¥ˆà¥¤", dialectProf.ToneHint)
	aiPrompt := stealthComposer.BuildPrompt(studentName, studentState, studentDistrict, sessionInfo.SessionPhase, string(currentTier), isHoliday, dialectProf)

	// 5. à¤¸à¤‚à¤ªà¥‚à¤°à¥à¤£ à¤¸à¤¿à¤¸à¥à¤Ÿà¤® à¤°à¤¿à¤ªà¥‹à¤°à¥à¤Ÿ
	fmt.Println("\n=======================================================")
	fmt.Printf("â±ï¸ à¤²à¤¾à¤‡à¤µ à¤¸à¤°à¥à¤µà¤° à¤Ÿà¤¾à¤‡à¤®à¤¸à¥à¤Ÿà¥ˆà¤®à¥à¤ª: %s (%s)\n", timeSnap.FormattedTimestamp, timeSnap.AcademicSessionLabel)
	fmt.Printf("ðŸ“… à¤•à¤² à¤•à¥€ à¤¤à¤¾à¤°à¥€à¤–: %s | à¤®à¤¾à¤¹ à¤•à¤¾ à¤…à¤‚à¤¤: %t | à¤µà¤°à¥à¤· à¤•à¤¾ à¤…à¤‚à¤¤: %t\n", timeSnap.NextDayDateString, timeSnap.IsMonthLastDay, timeSnap.IsYearLastDay)
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("âœ… à¤¸à¤¬à¤®à¤¿à¤¶à¤¨ à¤¸à¥à¤¥à¤¿à¤¤à¤¿: %s (à¤¶à¥‡à¤¯à¤°à¤¿à¤‚à¤— à¤°à¤¿à¤¸à¥à¤•: 0%%)\n", userMsg)
	fmt.Printf("ðŸ‘¤ à¤ªà¥à¤²à¤¾à¤¨: %s (à¤µà¥ˆà¤§à¤¤à¤¾: %d à¤¦à¤¿à¤¨, à¤•à¥‹à¤Ÿà¤¾: %d à¤¸à¥à¤•à¥ˆà¤¨/à¤¦à¤¿à¤¨)\n", currentTier, limits.TrialDays, limits.MaxDailyScans)
	fmt.Printf("ðŸ“ à¤°à¤¾à¤œà¥à¤¯-à¤µà¤¾à¤° à¤¸à¤¤à¥à¤°: %s\n", sessionInfo.SessionPhase)
	fmt.Printf("â„ï¸ à¤°à¤¾à¤œà¥à¤¯-à¤µà¤¾à¤° à¤µà¤¿à¤‚à¤Ÿà¤° à¤¸à¥à¤¥à¤¿à¤¤à¤¿: %t (à¤¶à¥‡à¤· à¤¦à¤¿à¤¨: %d)\n", winterInfo.IsWinterBreak, winterInfo.DaysLeft)
	fmt.Printf("ðŸŽ¯ à¤ªà¤°à¥€à¤•à¥à¤·à¤¾ à¤®à¥‹à¤¡: %s\n", examStatus.ExamHeadline)
	fmt.Printf("ðŸŒŸ à¤¬à¤¾à¤²-à¤¸à¥à¤µà¤°à¥à¤šà¤¿ à¤µà¤¿à¤·à¤¯: \"%s\"\n", childTrack.TopicName)
	fmt.Printf("ðŸ—£ï¸ à¤¸à¤•à¥à¤°à¤¿à¤¯ à¤¬à¥‹à¤²à¥€: %s (%s)\n", dialectProf.DialectCode, dialectProf.RegionHint)
	fmt.Printf("ðŸ“… à¤…à¤µà¤•à¤¾à¤¶ à¤¸à¥à¤¥à¤¿à¤¤à¤¿: %t (%s, à¤¶à¥‡à¤· à¤¦à¤¿à¤¨: %d)\n", isHoliday, holidayType, daysLeft)
	fmt.Printf("ðŸ’° à¤¨à¤µà¥€à¤¨à¥€à¤•à¤°à¤£ à¤¸à¥à¤¥à¤¿à¤¤à¤¿: â‚¹%.2f (%s)\n", renewalFee, loyaltyMsg)
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("ðŸŽ™ï¸ à¤µà¥‰à¤‡à¤¸ SSML:\n%s\n", kidSSML)
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("ðŸ¤– à¤®à¤¾à¤¸à¥à¤Ÿà¤° à¤¸à¥à¤Ÿà¥€à¤²à¥à¤¥ AI à¤ªà¥à¤°à¥‰à¤®à¥à¤ªà¥à¤Ÿ:\n%s\n", aiPrompt)
	fmt.Println("=======================================================")

	_ = demoService
	_ = homeworkDecomposer
	_ = foundationBridge
	_ = pacingService
	_ = panIndiaDialect
	_ = supportEngine
	_ = geoGuard
	_ = mindReader
	_ = inactivityNudge
}
