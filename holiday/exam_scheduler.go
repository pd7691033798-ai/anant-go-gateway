package holiday

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ExamMode string

const (
	ModeRegular            ExamMode = "REGULAR_STUDY"
	ModeClassTestTomorrow  ExamMode = "CLASS_TEST_TOMORROW"
	ModeHalfYearlyRevision ExamMode = "HALF_YEARLY_PREP"
	ModeFinalExamSprint    ExamMode = "FINAL_EXAM_SPRINT"
)

type ExamStatus struct {
	ActiveMode   ExamMode
	Subject      string
	DaysToExam   int
	ExamHeadline string
}

type ExamSchedulerService struct {
	db *sql.DB
}

func NewExamSchedulerService(db *sql.DB) *ExamSchedulerService {
	return &ExamSchedulerService{db: db}
}

// GetActiveExamMode: यूज़र के मैसेज या डेटाबेस शेड्यूल से एक्टिव एग्जाम मोड निकालता है
func (e *ExamSchedulerService) GetActiveExamMode(phone, userRawText string) ExamStatus {
	lower := strings.ToLower(userRawText)

	// 1. यदि छात्र या अभिभावक सीधे मैसेज में टेस्ट की सूचना दें
	if strings.Contains(lower, "कल टेस्ट") || strings.Contains(lower, "काल टेस्ट") || strings.Contains(lower, "कल पेपर") {
		return ExamStatus{
			ActiveMode:   ModeClassTestTomorrow,
			Subject:      "सामान्य विषय",
			DaysToExam:   1,
			ExamHeadline: "🔥 कल का टेस्ट: 15-मिनट रैपिड फॉर्मूला & टॉप प्रश्न अभ्यास",
		}
	}

	if e.db == nil {
		return ExamStatus{
			ActiveMode:   ModeRegular,
			Subject:      "नियमित",
			DaysToExam:   0,
			ExamHeadline: "नियमित 15-मिनट अभ्यास",
		}
	}

	// 2. डेटाबेस से शेड्यूल की जांच
	today := time.Now().Format("2006-01-02")
	var examType, subject string
	var startDate time.Time

	query := `SELECT exam_type, subject, start_date FROM student_exam_schedules 
	          WHERE student_phone = $1 AND is_active = TRUE AND end_date >= $2 
	          ORDER BY start_date ASC LIMIT 1`

	err := e.db.QueryRow(query, phone, today).Scan(&examType, &subject, &startDate)
	if err == nil {
		daysLeft := int(startDate.Sub(time.Now()).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}

		if examType == "HALF_YEARLY" {
			return ExamStatus{
				ActiveMode:   ModeHalfYearlyRevision,
				Subject:      subject,
				DaysToExam:   daysLeft,
				ExamHeadline: fmt.Sprintf("📚 हाफ ईयरली स्पेशल (शेष %d दिन): चैप्टर-वाइज़ स्पीड रिवीजन", daysLeft),
			}
		} else if examType == "FINAL_EXAM" {
			return ExamStatus{
				ActiveMode:   ModeFinalExamSprint,
				Subject:      subject,
				DaysToExam:   daysLeft,
				ExamHeadline: fmt.Sprintf("🏆 वार्षिक परीक्षा स्प्रिंट (शेष %d दिन): बोर्ड/फाइनल स्टेप-मार्किंग", daysLeft),
			}
		}
	}

	return ExamStatus{
		ActiveMode:   ModeRegular,
		Subject:      "नियमित",
		DaysToExam:   0,
		ExamHeadline: "नियमित 15-मिनट अभ्यास",
	}
}

// BuildExamAIPrompt: AI इंजन को परीक्षा मोड के अनुसार प्रॉम्प्ट तैयार करके देता है
func (e *ExamSchedulerService) BuildExamAIPrompt(status ExamStatus) string {
	switch status.ActiveMode {
	case ModeClassTestTomorrow:
		return "=== परीक्षा मोड (कल क्लास टेस्ट है) ===\n" +
			"1. छात्र को कोई नया थ्योरी पाठ न दें।\n" +
			"2. कल के टेस्ट के लिए 3 सबसे महत्वपूर्ण संभावित प्रश्न और फॉर्मूला लिखने को कहें।"
	case ModeHalfYearlyRevision:
		return fmt.Sprintf("=== हाफ-इयरली रिवीजन मोड (शेष %d दिन) ===\n"+
			"1. %s विषय के मुख्य अध्यायों का संक्षिप्त 15-मिनट लिखित रिवीजन कराएं।", status.Subject)
	case ModeFinalExamSprint:
		return fmt.Sprintf("=== वार्षिक परीक्षा फाइनल स्प्रिंट (शेष %d दिन) ===\n"+
			"1. फाइनल परीक्षा के अंक वितरण (Step-Marking) के अनुसार 1-1 प्रश्न की सघन चेकिंग करें।", status.DaysToExam)
	default:
		return "नियमित शैक्षणिक मूल्यांकन जारी रखें।"
	}
}

// ActivateIntensiveRevision (P14): अल्टीमेट फैमिली प्लान के लिए 1 से 7-दिवसीय इंटेंसिव रिवीजन ट्रिगर
func (ess *ExamSchedulerService) ActivateIntensiveRevision(phone string, examName string, days int) string {
	if days <= 0 || days > 7 {
		days = 7
	}
	return fmt.Sprintf("🎯 *%s इंटेंसिव रिवीजन मोड सक्रिय (Ultimate Family)!*\n• अवधि: अगले %d दिन\n• 4 छात्रों की संयुक्त प्रोग्रेस व पिछले वर्षों के पेपर्स पर फोकस।", examName, days)
}
