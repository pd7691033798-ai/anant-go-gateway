package vacation

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

type StudentCurriculumState struct {
	PhoneNumber      string
	CurrentGrade     int
	AdvanceGrade     int
	IsSick           bool
	MissedTopics     []string
	LastBookScanDate time.Time
	HasScannedBook   bool
	LifetimeHistory  []string
}

type PacingService struct {
	db       *sql.DB
	students map[string]*StudentCurriculumState
	mu       sync.RWMutex
}

func NewPacingService() *PacingService {
	return &PacingService{
		students: make(map[string]*StudentCurriculumState),
	}
}

// MarkStudentSick: बीमारी की स्थिति में तारीख री-शेड्यूल करना (पॉइंट 9)
func (ps *PacingService) MarkStudentSick(phone string, missedTopic string) string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	st := ps.getOrCreate(phone)
	st.IsSick = true
	st.MissedTopics = append(st.MissedTopics, missedTopic)

	return fmt.Sprintf("🩺 *स्वास्थ्य प्राथमिकता:* तबियत खराब होने के कारण टॉपिक '%s' को आगे री-शेड्यूल कर दिया गया है। स्वस्थ होने पर सिस्टम यह टॉपिक याद करवाकर नया टेस्ट लेगा।", missedTopic)
}

// GetDailyTopicWithBacklog: छूटा हुआ टॉपिक पहले पढ़ाना, फिर आज का कार्य (पॉइंट 10)
func (ps *PacingService) GetDailyTopicWithBacklog(phone string, todayTopic string) (string, []string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	st := ps.getOrCreate(phone)
	if len(st.MissedTopics) > 0 {
		backlog := st.MissedTopics[0]
		st.MissedTopics = st.MissedTopics[1:] // पिछला टॉपिक निकाला
		return fmt.Sprintf("📚 *बैकलॉग रिवीज़न:* पहले छूटा हुआ टॉपिक '%s' पूरा करें, उसके बाद आज का टॉपिक '%s' शुरू होगा।", backlog, todayTopic), []string{backlog, todayTopic}
	}

	return fmt.Sprintf("📖 आज का दैनिक टॉपिक: *%s*", todayTopic), []string{todayTopic}
}

// ValidateBookScan: बुक स्कैन साल में केवल एक बार नए सेशन में (पॉइंट 11)
func (ps *PacingService) ValidateBookScan(phone string) (bool, string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	st := ps.getOrCreate(phone)
	oneYearAgo := time.Now().AddDate(-1, 0, 0)

	if st.HasScannedBook && st.LastBookScanDate.After(oneYearAgo) {
		return false, "ℹ️ *बुक स्कैन सीमा:* आपकी किताब इस शैक्षणिक सत्र के लिए पहले से स्कैन और एक्टिव है। नया स्कैन केवल अगले साल नए सत्र में होगा।"
	}

	st.HasScannedBook = true
	st.LastBookScanDate = time.Now()
	return true, "✅ *सत्र बुक स्कैन स्वीकृत:* पूरे 1 वर्ष के लिए आपकी पाठ्यपुस्तक सिस्टम में सफलतापूर्वक लोड हो गई है।"
}

// ActivateVacationBridgeCourse: वेकेशन मोड में कक्षा 5 के बच्चे को कक्षा 6 का अग्रिम अध्ययन (पॉइंट 13)
func (ps *PacingService) ActivateVacationBridgeCourse(phone string, currentGrade int) string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	st := ps.getOrCreate(phone)
	st.CurrentGrade = currentGrade
	st.AdvanceGrade = currentGrade + 1

	return fmt.Sprintf("🏖️ *Vacation Mode (फाउंडेशन ब्रिज सक्रिय):*\nकक्षा %d की छुट्टियों में छात्र को कक्षा %d के मुख्य कॉन्सेप्ट्स और बेसिक्स का अग्रिम अध्ययन कराया जाएगा!",
		st.CurrentGrade, st.AdvanceGrade)
}

func (ps *PacingService) getOrCreate(phone string) *StudentCurriculumState {
	if s, ok := ps.students[phone]; ok {
		return s
	}
	s := &StudentCurriculumState{
		PhoneNumber:     phone,
		CurrentGrade:    5,
		AdvanceGrade:    6,
		MissedTopics:    make([]string, 0),
		LifetimeHistory: make([]string, 0),
	}
	ps.students[phone] = s
	return s
}
