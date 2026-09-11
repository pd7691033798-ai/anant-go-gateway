package vacation

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// HandwritingTelemetry हस्तलेखन और व्यवहार के मेट्रिक्स
type HandwritingTelemetry struct {
	AverageStrokeStability float64 // अक्षरों की स्थिरता (0.0 से 1.0)
	FatigueErrorRate       float64 // थकान या एकाग्रता टूटने की दर
	IsBedWritingAngle      bool    // क्या कॉपी लेटे-लेटे या असामान्य कोण पर लिखी गई है
	IncompleteSubmission   bool    // क्या कार्य बीच में ही छोड़ दिया गया
}

// StudentCurriculumState छात्र की शैक्षणिक व स्वास्थ्य स्थिति
type StudentCurriculumState struct {
	PhoneNumber            string
	CurrentGrade           int
	AdvanceGrade           int
	IsSick                 bool
	SickReportDate         time.Time
	MissedTopics           []string
	LastBookScanDate       time.Time
	HasScannedBook         bool
	LifetimeHistory        []string
	RecentStabilityHistory []float64 // पिछले 7 दिनों का बेसलाइन औसत
}

// ChildWellnessService बच्चे के स्वास्थ्य, थकान और पढ़ाई के संतुलन का मुख्य इंजन
type ChildWellnessService struct {
	db       *sql.DB
	students map[string]*StudentCurriculumState
	mu       sync.RWMutex
}

func NewChildWellnessService(db *sql.DB) *ChildWellnessService {
	return &ChildWellnessService{
		db:       db,
		students: make(map[string]*StudentCurriculumState),
	}
}

// AnalyzeHandwritingHealth: हस्तलेखन और विजुअल मैट्रिक्स से बीमारी/थकान का स्वतः आंकलन
func (cs *ChildWellnessService) AnalyzeHandwritingHealth(phone string, studentName string, metrics HandwritingTelemetry) (isSuspectedSick bool, alertMessage string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	st := cs.getOrCreate(phone)

	// बेसलाइन औसत निकालना
	var baseline float64 = 0.85
	if len(st.RecentStabilityHistory) > 0 {
		var total float64
		for _, v := range st.RecentStabilityHistory {
			total += v
		}
		baseline = total / float64(len(st.RecentStabilityHistory))
	}

	// 1. यदि अक्षरों में 30% से अधिक विचलन (Tremor) और थकान की गलतियाँ हों
	stabilityDrop := baseline - metrics.AverageStrokeStability
	isSevereTremor := stabilityDrop > 0.30
	isFatigued := metrics.FatigueErrorRate > 0.40 || metrics.IncompleteSubmission

	if (isSevereTremor && isFatigued) || (metrics.IsBedWritingAngle && isSevereTremor) {
		alert := fmt.Sprintf(
			"🌸 *अनंत अभ्यास — आत्मीय शिक्षक संदेश*\n\n"+
				"नमस्कार जी,\n"+
				"आज *%s* की कॉपी में अक्षरों का खिंचाव और लिखावट हमेशा जैसी सामान्य नहीं दिख रही है। ऐसा प्रतीत होता है कि बच्चा अस्वस्थ है या अत्यधिक थका हुआ है।\n\n"+
				"• आज के शेष अभ्यास को रोक दिया गया है ताकि बच्चे पर मानसिक दबाव न बने।\n"+
				"• कृपया पहले बच्चे का स्वास्थ्य और तापमान देखें।\n"+
				"• यदि बच्चा सचमुच अस्वस्थ है, तो केवल *'बीमार'* लिख दें—हम टॉपिक को स्वस्थ होने तक आगे री-शेड्यूल कर देंगे।\n\n"+
				"स्वास्थ्य ही प्रथम प्राथमिकता है! 🩺",
			studentName,
		)
		return true, alert
	}

	// सामान्य दिन: टेलीमेट्री हिस्ट्री अपडेट करें (अधिकतम 7 रिकॉर्ड)
	if len(st.RecentStabilityHistory) >= 7 {
		st.RecentStabilityHistory = st.RecentStabilityHistory[1:]
	}
	st.RecentStabilityHistory = append(st.RecentStabilityHistory, math.Max(0.1, metrics.AverageStrokeStability))

	return false, ""
}

// DetectSicknessFromMessage: चैट में अभिभावक द्वारा लिखे गए शब्दों की पहचान
func (cs *ChildWellnessService) DetectSicknessFromMessage(text string) bool {
	t := strings.ToLower(strings.TrimSpace(text))
	keywords := []string{
		"बीमार", "तबीयत", "तबीयत खराब", "बुखार", "hospital",
		"sick", "unwell", "fever", "डॉक्टर", "दवा", "थका",
	}
	for _, kw := range keywords {
		if strings.Contains(t, kw) {
			return true
		}
	}
	return false
}

// MarkStudentSick: बीमारी दर्ज करना और अभिभावक को आश्वस्त करने वाला संदेश तैयार करना
func (cs *ChildWellnessService) MarkStudentSick(phone string, studentName string, missedTopic string) string {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	st := cs.getOrCreate(phone)
	st.IsSick = true
	st.SickReportDate = time.Now()
	if missedTopic != "" {
		st.MissedTopics = append(st.MissedTopics, missedTopic)
	}

	return fmt.Sprintf(
		"🩺 *अनंत अभ्यास - स्वास्थ्य प्राथमिकता*\n\n"+
			"आदरणीय अभिभावक जी, हमने *%s* की अस्वस्थता दर्ज कर ली है।\n"+
			"• आज का टॉपिक (*%s*) सुरक्षित रूप से आगे के लिए री-शेड्यूल कर दिया गया है।\n"+
			"• बच्चे पर पढ़ाई का कोई तनाव न दें।\n"+
			"• स्वस्थ होने पर बस *'अब ठीक है'* लिखें या सीधे कॉपी स्कैन करें, सिस्टम वहीं से अभ्यास शुरू करा देगा।\n\n"+
			"ईश्वर से बच्चे के शीघ्र स्वास्थ्य लाभ की प्रार्थना! 🌸",
		studentName, missedTopic,
	)
}

// MarkStudentRecovered: बच्चे के ठीक होने पर पुनः अभ्यास चालू करना
func (cs *ChildWellnessService) MarkStudentRecovered(phone string, studentName string) string {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	st := cs.getOrCreate(phone)
	st.IsSick = false

	return fmt.Sprintf(
		"🎉 *स्वागत है %s!*\n\n"+
			"यह जानकर बहुत प्रसन्नता हुई कि आप पूर्णतः स्वस्थ हो चुके हैं।\n"+
			"आपका पिछला बैकलॉग सुरक्षित है। आज से हमारा 15 मिनट का नियमित अभ्यास पुनः शुरू होता है! 📚✍️",
		studentName,
	)
}

// GetDailyTopicWithBacklog: छूटा हुआ टॉपिक पहले कराना, फिर नया कार्य
func (cs *ChildWellnessService) GetDailyTopicWithBacklog(phone string, todayTopic string) (string, []string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	st := cs.getOrCreate(phone)
	if len(st.MissedTopics) > 0 {
		backlog := st.MissedTopics[0]
		st.MissedTopics = st.MissedTopics[1:]
		return fmt.Sprintf("📚 *बैकलॉग रिवीज़न:* पहले छूटा हुआ टॉपिक '%s' पूरा करें, उसके बाद आज का टॉपिक '%s' शुरू होगा।", backlog, todayTopic), []string{backlog, todayTopic}
	}

	return fmt.Sprintf("📖 आज का दैनिक टॉपिक: *%s*", todayTopic), []string{todayTopic}
}

// ValidateBookScan: पूरे 1 वर्ष के लिए पाठ्यपुस्तक केवल एक बार मान्य
func (cs *ChildWellnessService) ValidateBookScan(phone string) (bool, string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	st := cs.getOrCreate(phone)
	oneYearAgo := time.Now().AddDate(-1, 0, 0)

	if st.HasScannedBook && st.LastBookScanDate.After(oneYearAgo) {
		return false, "ℹ️ *बुक स्कैन सीमा:* आपकी पाठ्यपुस्तक इस शैक्षणिक सत्र के लिए पहले से सक्रिय है।"
	}

	st.HasScannedBook = true
	st.LastBookScanDate = time.Now()
	return true, "✅ *सत्र बुक स्कैन स्वीकृत:* पूरे 1 वर्ष के लिए आपकी पाठ्यपुस्तक सिस्टम में सुरक्षित लोड हो गई है।"
}

// ActivateVacationBridgeCourse: छुट्टियों में अगली कक्षा का अग्रिम फाउंडेशन
func (cs *ChildWellnessService) ActivateVacationBridgeCourse(phone string, currentGrade int) string {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	st := cs.getOrCreate(phone)
	st.CurrentGrade = currentGrade
	st.AdvanceGrade = currentGrade + 1

	return fmt.Sprintf("🏖️ *Vacation Mode (फाउंडेशन ब्रिज सक्रिय):*\nकक्षा %d की छुट्टियों में छात्र को कक्षा %d के मुख्य कॉन्सेप्ट्स का अग्रिम अध्ययन कराया जाएगा!",
		st.CurrentGrade, st.AdvanceGrade)
}

func (cs *ChildWellnessService) getOrCreate(phone string) *StudentCurriculumState {
	if s, ok := cs.students[phone]; ok {
		return s
	}
	s := &StudentCurriculumState{
		PhoneNumber:            phone,
		CurrentGrade:           5,
		AdvanceGrade:           6,
		MissedTopics:           make([]string, 0),
		LifetimeHistory:        make([]string, 0),
		RecentStabilityHistory: []float64{0.85, 0.88, 0.84},
	}
	cs.students[phone] = s
	return s
}
