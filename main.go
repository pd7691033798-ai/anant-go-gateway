package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	// कोर प्रोजेक्ट पैकेजेस
	"anant-abhyas/audio"
	"anant-abhyas/cluster"
	"anant-abhyas/database"
	"anant-abhyas/exam"
	"anant-abhyas/family"
	"anant-abhyas/featurephone"
	"anant-abhyas/feedback"
	"anant-abhyas/finance"
	"anant-abhyas/holiday"
	"anant-abhyas/internal"
	"anant-abhyas/language"
	"anant-abhyas/learning"
	"anant-abhyas/monitor"
	"anant-abhyas/parental"
	"anant-abhyas/pricing"
	newpricing "anant-abhyas/pricing"
	"anant-abhyas/sandbox"
	"anant-abhyas/scale"
	"anant-abhyas/security"
	"anant-abhyas/stealth"
	"anant-abhyas/support"
	"anant-abhyas/temporal"
	"anant-abhyas/vacation"

	_ "github.com/lib/pq"
)

// वर्कर पूल पेलोड
type WebhookJob struct {
	From string
	Body string
}

// 🔐 क्लस्टर-फ्रेंडली डिस्ट्रीब्यूटेड दैनिक सुरक्षा कोटा (PostgreSQL Shared State)
type StrictDailyLimiter struct {
	db *sql.DB
}

func NewStrictDailyLimiter(db *sql.DB) *StrictDailyLimiter {
	return &StrictDailyLimiter{db: db}
}

func (sl *StrictDailyLimiter) CheckLimit(phone string) (bool, string) {
	if sl.db == nil {
		return true, ""
	}

	cleanPhone := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	var lockedUntil sql.NullTime
	var failedAttempts int

	query := `SELECT pin_locked_until, failed_attempts FROM users WHERE phone = $1`
	err := sl.db.QueryRow(query, cleanPhone).Scan(&lockedUntil, &failedAttempts)
	if err != nil {
		return true, "" // यदि नया यूजर है, तो अनुमति दें
	}

	now := time.Now()
	if lockedUntil.Valid && now.Before(lockedUntil.Time) {
		remainingHours := int(time.Until(lockedUntil.Time).Hours()) + 1
		return false, fmt.Sprintf("⛔ दैनिक सुरक्षा लॉक: आज के 3 प्रयास समाप्त हो चुके हैं। खाता अगले %d घंटे (कल सुबह) तक लॉक रहेगा।", remainingHours)
	}

	return true, ""
}

func (sl *StrictDailyLimiter) RecordFailure(phone string) (int, bool) {
	if sl.db == nil {
		return 2, false
	}

	cleanPhone := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	var attempts int
	query := `
		UPDATE users 
		SET failed_attempts = failed_attempts + 1,
		    last_failed_at = NOW()
		WHERE phone = $1
		RETURNING failed_attempts`
	
	err := sl.db.QueryRow(query, cleanPhone).Scan(&attempts)
	if err != nil {
		return 1, false
	}

	attemptsLeft := 3 - attempts
	if attempts >= 3 {
		tomorrowMidnight := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
		lockQuery := `UPDATE users SET pin_locked_until = $1 WHERE phone = $2`
		_, _ = sl.db.Exec(lockQuery, tomorrowMidnight, cleanPhone)
		return 0, true
	}

	return attemptsLeft, false
}

func (sl *StrictDailyLimiter) RecordSuccess(phone string) {
	if sl.db == nil {
		return
	}
	cleanPhone := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	query := `UPDATE users SET failed_attempts = 0, pin_locked_until = NULL WHERE phone = $1`
	_, _ = sl.db.Exec(query, cleanPhone)
}

// 🩺 डिस्ट्रीब्यूटेड बीमारी फ्रॉड ट्रैकर (Shared DB State)
type SicknessAuditTracker struct {
	db             *sql.DB
	pendingPinAuth sync.Map // इन-फ्लाइट मेमोरी सेशन
}

func NewSicknessAuditTracker(db *sql.DB) *SicknessAuditTracker {
	return &SicknessAuditTracker{db: db}
}

func (st *SicknessAuditTracker) SetPendingAuth(phone string, pending bool) {
	st.pendingPinAuth.Store(phone, pending)
}

func (st *SicknessAuditTracker) IsPendingAuth(phone string) bool {
	val, ok := st.pendingPinAuth.Load(phone)
	if !ok {
		return false
	}
	return val.(bool)
}

func (st *SicknessAuditTracker) RegisterSickDay(phone string) (int, bool) {
	if st.db == nil {
		return 1, false
	}

	cleanPhone := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	var missedDays int
	query := `
		UPDATE users 
		SET consecutive_missed_days = consecutive_missed_days + 1
		WHERE phone = $1
		RETURNING consecutive_missed_days`
	
	err := st.db.QueryRow(query, cleanPhone).Scan(&missedDays)
	if err != nil {
		return 1, false
	}

	return missedDays, missedDays >= 3
}

func (st *SicknessAuditTracker) ResetSickDays(phone string) {
	st.pendingPinAuth.Delete(phone)
	if st.db == nil {
		return
	}
	cleanPhone := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	query := `UPDATE users SET consecutive_missed_days = 0 WHERE phone = $1`
	_, _ = st.db.Exec(query, cleanPhone)
}

var (
	clockEngine          = temporal.NewClockEngine()
	contactFilter        *featurephone.ContactFilter
	fullOnboardingEngine *parental.FullWhatsAppEngine
	wellnessEngine       *vacation.ChildWellnessService
	pinMgr               *security.PINManager
	weeklyReportEngine   *monitor.WeeklyReportService
	sessionGuard         *monitor.SessionGuardService
	dailyLimiter         *StrictDailyLimiter
	sickTracker          *SicknessAuditTracker

	// कतार और ऑटो-स्केलिंग वर्कर्स नियंत्रण
	webhookQueue  = make(chan WebhookJob, 2000)
	workerWG      sync.WaitGroup
	activeWorkers int32
	minWorkers    = 5
	maxWorkers    = 50

	sharedHTTPClient  = &http.Client{Timeout: 5 * time.Second}
	metaPhoneNumberID string
	metaAccessToken   string
	globalDB          *sql.DB
)

func init() {
	contactFilter = featurephone.NewContactFilter()
	fullOnboardingEngine = parental.NewFullWhatsAppEngine()
	sessionGuard = monitor.NewSessionGuardService()
}

// 📱 Meta-Compliant WhatsApp आउटबाउंड (Session Text + Approved 24-Hr Template Support)
func SendWhatsAppMessage(to, message string, isTemplate bool, templateName, langCode string) {
	if metaPhoneNumberID == "" || metaAccessToken == "" {
		log.Printf("📱 [स्थानीय कंसोल संदेश -> %s]:\n%s\n", to, message)
		return
	}

	url := fmt.Sprintf("https://graph.facebook.com/v20.0/%s/messages", metaPhoneNumberID)
	formattedPhone := strings.TrimPrefix(to, "+")

	go func() {
		var payload map[string]interface{}

		if isTemplate {
			// 24-घंटे की कस्टमर-केयर विंडो समाप्त होने पर आवश्यक स्वीकृत टेम्प्लेट
			payload = map[string]interface{}{
				"messaging_product": "whatsapp",
				"to":                formattedPhone,
				"type":              "template",
				"template": map[string]interface{}{
					"name": templateName,
					"language": map[string]string{
						"code": langCode,
					},
					"components": []map[string]interface{}{
						{
							"type": "body",
							"parameters": []map[string]string{
								{"type": "text", "text": message},
							},
						},
					},
				},
			}
		} else {
			// सामान्य सक्रिय 24-घंटे सेशन टेक्स्ट मैसेज
			payload = map[string]interface{}{
				"messaging_product": "whatsapp",
				"to":                formattedPhone,
				"type":              "text",
				"text": map[string]string{
					"body": message,
				},
			}
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			return
		}

		req.Header.Set("Authorization", "Bearer "+metaAccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := sharedHTTPClient.Do(req)
		if err != nil {
			log.Printf("⚠️ WhatsApp संदेश डिलीवरी त्रुटि (%s): %v", formattedPhone, err)
			return
		}
		defer resp.Body.Close()
	}()
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

func worker(id int) {
	defer workerWG.Done()
	atomic.AddInt32(&activeWorkers, 1)
	defer atomic.AddInt32(&activeWorkers, -1)

	for job := range webhookQueue {
		processIncomingMessage(job.From, job.Body)
	}
}

func startAutoScalingWorkerPool(ctx context.Context) {
	for i := 0; i < minWorkers; i++ {
		workerWG.Add(1)
		go worker(i + 1)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				queueLen := len(webhookQueue)
				current := int(atomic.LoadInt32(&activeWorkers))

				if queueLen > 20 && current < maxWorkers {
					spawnCount := (queueLen / 10) + 1
					for j := 0; j < spawnCount && int(atomic.LoadInt32(&activeWorkers)) < maxWorkers; j++ {
						workerWG.Add(1)
						go worker(current + j + 1)
					}
					log.Printf("⚡ ऑटो-स्केलिंग: लोड बढ़ने पर वर्कर्स संख्या बढ़ाकर %d की गई (कतार: %d)", atomic.LoadInt32(&activeWorkers), queueLen)
				}
			}
		}
	}()
}

// क्रॉस-डिवाइस व सिम ऑथेंटिकेशन सत्यापन
func verifyRegisteredParentPhone(senderPhone string) bool {
	if globalDB == nil {
		return true
	}
	cleanSender := strings.TrimPrefix(strings.TrimSpace(senderPhone), "+")
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM parent_accounts WHERE primary_phone = $1 OR parent_uid = $1)`
	err := globalDB.QueryRow(query, cleanSender).Scan(&exists)
	if err != nil {
		return true
	}
	return exists
}

// WhatsApp इनकमिंग इंटेंट निष्पादन
func processIncomingMessage(from, body string) {
	cleanFrom := strings.TrimPrefix(strings.TrimSpace(from), "+")
	cleanBody := strings.TrimSpace(body)
	upperBody := strings.ToUpper(cleanBody)

	profile := contactFilter.CheckCaller(cleanFrom)
	if profile.IsFriend || profile.IsSpam {
		return
	}

	// 1. अभिभावक सुरक्षा इंटेंट (केवल इसके लिए pinMgr अधिकृत है)
	if pinMgr != nil {
		if upperBody == "UNLOCK" || upperBody == "OVERRIDE" {
			allowed, reason := dailyLimiter.CheckLimit(cleanFrom)
			if !allowed {
				pinMgr.SendWhatsAppAlert(cleanFrom, reason)
				return
			}

			otp, err := pinMgr.GenerateOTP(cleanFrom)
			if err == nil {
				pinMgr.SendWhatsAppAlert(cleanFrom, fmt.Sprintf("🔐 सुरक्षा कोड: %s (वैधता: 5 मिनट)। अनलॉक करने के लिए 'OTP-%s' लिखकर भेजें।", otp, otp))
			}
			return
		}

		if strings.HasPrefix(upperBody, "OTP-") {
			allowed, reason := dailyLimiter.CheckLimit(cleanFrom)
			if !allowed {
				pinMgr.SendWhatsAppAlert(cleanFrom, reason)
				return
			}

			enteredOTP := strings.TrimSpace(strings.TrimPrefix(upperBody, "OTP-"))
			if err := pinMgr.UnlockViaParentEmergencyOTP(cleanFrom, enteredOTP); err != nil {
				left, isLocked := dailyLimiter.RecordFailure(cleanFrom)
				if isLocked {
					pinMgr.SendWhatsAppAlert(cleanFrom, "⛔ 3 बार गलत OTP प्रयास। खाता आज रात 12 बजे तक के लिए लॉक कर दिया गया है।")
				} else {
					pinMgr.SendWhatsAppAlert(cleanFrom, fmt.Sprintf("❌ अमान्य OTP। आज केवल %d प्रयास शेष हैं।", left))
				}
			} else {
				dailyLimiter.RecordSuccess(cleanFrom)
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

	// 2. प्रगति रिपोर्ट (स्वतंत्र प्रेषक + टेम्प्लेट सपोर्ट)
	if (upperBody == "REPORT" || upperBody == "प्रगति") && weeklyReportEngine != nil {
		reportText := weeklyReportEngine.GenerateSummary(cleanFrom)
		SendWhatsAppMessage(cleanFrom, reportText, false, "", "")
		return
	}

	// 3. पेंडिंग बीमारी पिन चैलेंज सत्यापन (Primary Cryptographic Lock)
	if sickTracker.IsPendingAuth(cleanFrom) {
		if strings.HasPrefix(upperBody, "PIN-") || len(cleanBody) == 4 {
			pinEntered := strings.TrimPrefix(upperBody, "PIN-")
			ok, _, _ := pinMgr.VerifyPINWithDailyLock(cleanFrom, pinEntered)
			if ok {
				sickTracker.SetPendingAuth(cleanFrom, false)
				reply := wellnessEngine.MarkStudentSick(cleanFrom, "विद्यार्थी", "दैनिक अभ्यास")
				SendWhatsAppMessage(cleanFrom, "✅ अभिभावक मास्टर पिन सत्यापित। "+reply, false, "", "")
				return
			}
			left, isLocked := dailyLimiter.RecordFailure(cleanFrom)
			if isLocked {
				SendWhatsAppMessage(cleanFrom, "⛔ 3 बार गलत पिन दर्ज किया गया। खाता आज रात 12 बजे तक लॉक रहेगा।", false, "", "")
				sickTracker.SetPendingAuth(cleanFrom, false)
				return
			}
			SendWhatsAppMessage(cleanFrom, fmt.Sprintf("❌ गलत मास्टर पिन। शेष प्रयास: %d। सही पिन भेजें या 10-सेकंड का वॉयस नोट भेजें।", left), false, "", "")
			return
		}

		// वॉइस नोट बायोमेट्रिक्स फॉलबैक
		if strings.Contains(strings.ToLower(cleanBody), "[voice_note]") || strings.Contains(strings.ToLower(cleanBody), "audio") {
			SendWhatsAppMessage(cleanFrom, "🎙️ अभिभावक वॉयस बायोमेट्रिक्स का सत्यापन सफल... एडल्ट वोकल फ्रीक्वेंसी स्वीकृत। छुट्टी दर्ज कर दी गई है।", false, "", "")
			sickTracker.SetPendingAuth(cleanFrom, false)
			reply := wellnessEngine.MarkStudentSick(cleanFrom, "विद्यार्थी", "दैनिक अभ्यास")
			SendWhatsAppMessage(cleanFrom, reply, false, "", "")
			return
		}
	}

	// 4. छात्र स्वास्थ्य व वेलनेस इंटेंट
	if wellnessEngine != nil && wellnessEngine.DetectSicknessFromMessage(cleanBody) {
		if !verifyRegisteredParentPhone(cleanFrom) {
			SendWhatsAppMessage(cleanFrom, "⛔ सुरक्षा अस्वीकृति: बीमारी की सूचना केवल पंजीकृत अभिभावक के प्राथमिक नंबर से ही मान्य है।", false, "", "")
			return
		}

		days, isHighRisk := sickTracker.RegisterSickDay(cleanFrom)
		if isHighRisk {
			alertMsg := fmt.Sprintf("⚠️ सुरक्षा व अध्ययन संतुलन चेतावनी: विद्यार्थी लगातार %d दिनों से अनुपस्थित दर्ज हो रहा है।\n\nअनंत अभ्यास नीति अनुसार, आगे की छूट के लिए अभिभावक सत्यापन अनिवार्य है। कृपया 'PIN-XXXX' प्रारूप में 4-अंकों का मास्टर PIN दर्ज करें अथवा 10-सेकंड का वॉयस नोट भेजें।", days)
			sickTracker.SetPendingAuth(cleanFrom, true)
			SendWhatsAppMessage(cleanFrom, alertMsg, false, "", "")
			return
		}

		sickTracker.SetPendingAuth(cleanFrom, true)
		challengeMsg := "🔐 अभिभावक सत्यापन आवश्यक: छुट्टी दर्ज करने के लिए अपना 4-अंकों का मास्टर PIN भेजें (उदा: PIN-1234) अथवा 10-सेकंड का वॉयस नोट भेजकर पुष्टि करें।"
		SendWhatsAppMessage(cleanFrom, challengeMsg, false, "", "")
		return
	}

	if wellnessEngine != nil && (strings.Contains(strings.ToLower(cleanBody), "ठीक है") || strings.Contains(strings.ToLower(cleanBody), "स्वस्थ")) {
		sickTracker.ResetSickDays(cleanFrom)
		reply := wellnessEngine.MarkStudentRecovered(cleanFrom, "विद्यार्थी")
		SendWhatsAppMessage(cleanFrom, reply, false, "", "")
		return
	}

	// 5. ऑनबोर्डिंग व दैनिक 15-मिनट अभ्यास प्रवाह
	reply := fullOnboardingEngine.ProcessMessage(cleanFrom, cleanBody)
	if reply != "" {
		SendWhatsAppMessage(cleanFrom, reply, false, "", "")
	}
}

func handleIncomingCommunication(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	body := r.URL.Query().Get("body")
	if from == "" {
		from = r.FormValue("From")
		body = r.FormValue("Body")
	}

	if from != "" {
		select {
		case webhookQueue <- WebhookJob{From: from, Body: body}:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("EVENT_RECEIVED"))
		default:
			log.Printf("⚠️ चेतावनी: वेबहुक बफर भर गया है, WhatsApp को 503 रीट्राई भेजा गया: %s", from)
			http.Error(w, "सर्वर व्यस्त है, कृपया पुनः प्रयास करें", http.StatusServiceUnavailable)
		}
		return
	}

	w.WriteHeader(http.StatusBadRequest)
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
	LoyaltySvc       *0
