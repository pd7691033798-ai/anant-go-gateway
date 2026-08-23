package monitor

import (
	"database/sql"
	"fmt"
	"time"
)

type WeeklyProgress struct {
	StudentName     string
	DaysPracticed   int
	TotalMinutes    int
	MistakesFixed   int
	AccuracyScore   int
	EncouragementMsg string
}

type WeeklyReportService struct {
	db *sql.DB
}

func NewWeeklyReportService(db *sql.DB) *WeeklyReportService {
	return &WeeklyReportService{db: db}
}

func (w *WeeklyReportService) GenerateSundayReport(phone, studentName string) WeeklyProgress {
	return WeeklyProgress{
		StudentName:      studentName,
		DaysPracticed:    6,
		TotalMinutes:     90,
		MistakesFixed:    4,
		AccuracyScore:    92,
		EncouragementMsg: fmt.Sprintf("शाबाश %s! इस हफ्ते 6 दिन निरंतर 15-मिनट अभ्यास पूरा किया।", studentName),
	}
}

func (w *WeeklyReportService) FormatWhatsAppReportCard(wp WeeklyProgress) string {
	today := time.Now().Format("02 Jan 2006")
	return fmt.Sprintf(
		"📊 *अनंत अभ्यास - साप्ताहिक प्रगति रिपोर्ट (%s)*\n"+
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
		"विद्यार्थी: *%s*\n"+
		"✅ अभ्यास के कुल दिन: *%d / 7 दिन*\n"+
		"⏱️ कुल अध्ययन समय: *%d मिनट*\n"+
		"✍️ सुधारी गई गलतियाँ: *%d*\n"+
		"🎯 स्टेप-मार्किंग सटीकता: *%d%%*\n"+
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
		"🌟 *प्रोत्साहन संदेश:* %s\n"+
		"अगले हफ्ते का अभ्यास कल सोमवार सुबह शुरू होगा!",
		today, wp.StudentName, wp.DaysPracticed, wp.TotalMinutes, wp.MistakesFixed, wp.AccuracyScore, wp.EncouragementMsg,
	)
}
