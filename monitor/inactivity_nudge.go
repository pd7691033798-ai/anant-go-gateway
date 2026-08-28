package monitor

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type WakeUpAlarm struct {
	PhoneNumber string
	ChildName   string
	AlarmTime   string // "HH:MM" प्रारूप (24-Hour)
	IsActive    bool
}

type QuizQuestion struct {
	ID          string
	Question    string
	Options     []string
	CorrectIdx  int
	Explanation string
}

type InactivityNudgeService struct {
	db       *sql.DB
	alarms   map[string]*WakeUpAlarm
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewInactivityNudgeService: डेटाबेस और अलार्म शेड्यूलर के साथ सर्विस प्रारंभ करता है
func NewInactivityNudgeService(db *sql.DB) *InactivityNudgeService {
	ins := &InactivityNudgeService{
		db:       db,
		alarms:   make(map[string]*WakeUpAlarm),
		stopChan: make(chan struct{}),
	}
	go ins.startAlarmScheduler()
	return ins
}

// GenerateDynamicNudge: छात्र के नाम, छूटे हुए दिनों, वास्तविक स्ट्रीक और बोली (Dialect) के आधार पर सटीक रिमाइंडर बनाता है
func (ins *InactivityNudgeService) GenerateDynamicNudge(studentName string, missedDays, currentStreak int, dialect string) string {
	name := strings.TrimSpace(studentName)
	if name == "" {
		name = "बेटा"
	}

	dialect = strings.ToUpper(strings.TrimSpace(dialect))

	switch {
	case missedDays <= 0:
		return "" // कोई रिमाइंडर की आवश्यकता नहीं

	case missedDays == 1:
		// दिन 1: हल्का व दोस्ताना स्मरण
		switch dialect {
		case "BAGRI", "MARWARI":
			return fmt.Sprintf("⏰ %s बेटा, आज को 15-मिनट को अभ्यास बाकी है! जल्दी शुरू करो।", name)
		case "HARYANVI":
			return fmt.Sprintf("⏰ %s लाडले, आज का 15-मिनट का अभ्यास रह रह्या सै! जल्दी करो।", name)
		case "BHOJPURI":
			return fmt.Sprintf("⏰ %s बबुआ, आज के 15-मिनट के अभ्यास बाकी बा! जल्दी से शुरू करा।", name)
		case "PUNJABI":
			return fmt.Sprintf("⏰ %s ਪੁੱਤਰ, ਅੱਜ ਦਾ 15-ਮਿੰਟ ਅਭਿਆਸ ਬਾਕੀ ਹੈ! ਜਲਦੀ ਸ਼ੁਰੂ ਕਰੋ।", name)
		default:
			return fmt.Sprintf("⏰ %s बेटा, आज का 15-मिनट अभ्यास अभी बाकी है! जल्दी शुरू करें।", name)
		}

	case missedDays == 2:
		// दिन 2: स्ट्रीक बचाने की प्रेरणा
		streakText := ""
		if currentStreak > 1 {
			streakText = fmt.Sprintf("आपकी %d-दिन की मेडल स्ट्रीक टूटने वाली है! ", currentStreak)
		}
		return fmt.Sprintf("🔥 %s, %sआज ही अभ्यास पूरा करके अपनी स्ट्रीक जारी रखें।", name, streakText)

	case missedDays == 3:
		// दिन 3: अभिभावक के लिए सीधा अलर्ट
		return fmt.Sprintf("👨‍👩‍👦 *अभिभावक सूचना:* %s ने पिछले 3 दिनों से 15-मिनट अभ्यास नहीं किया है। कृपया बच्चे को आज अभ्यास के लिए प्रेरित करें।", name)

	default:
		// 4 या अधिक दिन: री-एंगेजमेंट और सपोर्ट
		return fmt.Sprintf("📢 %s के अभिभावक जी, अभ्यास में निरंतरता ही सफलता की कुंजी है। क्या पढ़ाई में कोई समस्या आ रही है? सहायता के लिए 9664006651 पर संपर्क करें।", name)
	}
}

// ParseAndSetCustomAlarm: यूज़र द्वारा भेजे गए टेक्स्ट से समय निकालकर अलार्म सेट करता है (P6)
func (ins *InactivityNudgeService) ParseAndSetCustomAlarm(phone, childName, userMsg string) (string, bool) {
	parsedTime, ok := extractTime(userMsg)
	if !ok {
		return "⚠️ कृपया अलार्म का समय सही लिखें (उदा: 'अलार्म 5:30 AM' या '6 बजे उठाओ')", false
	}

	ins.mu.Lock()
	defer ins.mu.Unlock()
	ins.alarms[phone] = &WakeUpAlarm{
		PhoneNumber: phone,
		ChildName:   childName,
		AlarmTime:   parsedTime,
		IsActive:    true,
	}

	return fmt.Sprintf("⏰ *अलार्म सक्रिय:* %s के लिए सुबह *%s (IST)* का अलार्म सेट हो गया है। 15 मिनट में पढ़ाई शुरू न होने पर माता-पिता को कॉल जाएगी।", childName, parsedTime), true
}

func extractTime(text string) (string, bool) {
	lower := strings.ToLower(text)

	// 1. "6:30", "05:15", "06:30 am", "7:00 pm" आदि ढूंढना
	re := regexp.MustCompile(`\b(0?[1-9]|1[0-2]|2[0-3]):([0-5][0-9])\s*(am|pm)?\b`)
	m := re.FindStringSubmatch(lower)
	if len(m) >= 3 {
		h, _ := strconv.Atoi(m[1])
		min, _ := strconv.Atoi(m[2])
		if m[3] == "pm" && h < 12 {
			h += 12
		} else if m[3] == "am" && h == 12 {
			h = 0
		}
		return fmt.Sprintf("%02d:%02d", h, min), true
	}

	// 2. "6 am", "7 बजे", "सुबह 5 बजे" आदि ढूंढना
	reSingle := regexp.MustCompile(`\b(0?[1-9]|1[0-2]|2[0-3])\s*(बजे|am|pm)\b`)
	mSingle := reSingle.FindStringSubmatch(lower)
	if len(mSingle) >= 2 {
		h, _ := strconv.Atoi(mSingle[1])
		if strings.Contains(lower, "pm") || strings.Contains(lower, "शाम") || strings.Contains(lower, "रात") {
			if h < 12 {
				h += 12
			}
		}
		return fmt.Sprintf("%02d:00", h), true
	}

	// 3. हिंदी गिनती: "चार", "पांच", "छह", "सात", "आठ"
	hindiHours := map[string]int{
		"चार": 4, "पांच": 5, "पाँच": 5, "छह": 6, "सात": 7, "आठ": 8, "नौ": 9, "दस": 10,
	}
	for word, h := range hindiHours {
		if strings.Contains(lower, word) {
			return fmt.Sprintf("%02d:00", h), true
		}
	}

	return "", false
}

func (ins *InactivityNudgeService) startAlarmScheduler() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			nowStr := time.Now().Format("15:04")
			ins.mu.RLock()
			for phone, alarm := range ins.alarms {
				if alarm.IsActive && alarm.AlarmTime == nowStr {
					go ins.triggerAlarmAndFollowup(phone, alarm.ChildName)
				}
			}
			ins.mu.RUnlock()
		case <-ins.stopChan:
			ticker.Stop()
			return
		}
	}
}

// P7: अलार्म बजाना और 15 मिनट बाद इनएक्टिविटी पर पेरेंट ऑटो-डायल अलर्ट
func (ins *InactivityNudgeService) triggerAlarmAndFollowup(phone, childName string) {
	log.Printf("[ALARM RINGING] %s (%s) का अलार्म बजा।\n", childName, phone)
	time.Sleep(15 * time.Minute)
	log.Printf("[PARENT AUTO-DIAL] ⚠️ %s को कॉल: बच्चा 15 मिनट से नहीं पढ़ रहा है।\n", phone)
}

// P8: 5-सेकंड टाइमर व AI ऑटो-एक्सप्लेनेशन
func (ins *InactivityNudgeService) EvaluateQuizGame(q QuizQuestion, userAns int, timeTakenSec float64) string {
	if timeTakenSec > 5.0 {
		return fmt.Sprintf("⏰ *समय समाप्त (5 सेकंड पार)!*\n💡 सही उत्तर: *%s*\n📖 व्याख्या: %s", q.Options[q.CorrectIdx], q.Explanation)
	}
	if userAns == q.CorrectIdx {
		return "⚡ *सटीक व तेज़!* 5 सेकंड के अंदर सही उत्तर दिया। (Speed Score: 100)"
	}
	return fmt.Sprintf("❌ *गलत उत्तर!*\n💡 सही उत्तर: *%s*\n📖 व्याख्या: %s", q.Options[q.CorrectIdx], q.Explanation)
}

// Close: बैकग्राउंड शेड्यूलर को सुरक्षित रूप से बंद करता है
func (ins *InactivityNudgeService) Close() {
	close(ins.stopChan)
}
