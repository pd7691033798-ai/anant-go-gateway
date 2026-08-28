package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type WeeklyProgress struct {
	StudentName      string
	DaysPracticed    int
	TotalMinutes     int
	MistakesFixed    int
	AccuracyScore    int
	EncouragementMsg string
}

type WeeklyReportService struct {
	db *sql.DB
}

func NewWeeklyReportService(db *sql.DB) *WeeklyReportService {
	return &WeeklyReportService{db: db}
}

// GenerateSundayReport पिछले 7 दिनों के वास्तविक अभ्यास डेटा की गणना करता है
func (w *WeeklyReportService) GenerateSundayReport(ctx context.Context, studentUID, studentName, dialect string) WeeklyProgress {
	name := strings.TrimSpace(studentName)
	if name == "" {
		name = "बेटा"
	}

	progress := WeeklyProgress{
		StudentName:   name,
		DaysPracticed: 0,
		TotalMinutes:  0,
		MistakesFixed: 0,
		AccuracyScore: 0,
	}

	if w.db != nil && strings.TrimSpace(studentUID) != "" {
		dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// पिछले 7 दिनों का वास्तविक डेटा: अभ्यास के दिन, कुल सेकंड, सुधारी गई गलतियाँ
		query := `
			SELECT 
				COALESCE(COUNT(DISTINCT DATE(created_at)), 0) AS days_practiced,
				COALESCE(SUM(time_taken_seconds), 0) / 60 AS total_minutes,
				COALESCE(SUM(CASE WHEN is_fixed = TRUE THEN 1 ELSE 0 END), 0) AS mistakes_fixed,
				COALESCE(AVG(score_percentage), 0) AS avg_accuracy
			FROM practice_sessions 
			WHERE student_uid = $1 
			  AND created_at >= NOW() - INTERVAL '7 days'
			  AND is_completed = TRUE`

		var avgAccuracy float64
		err := w.db.QueryRowContext(dbCtx, query, studentUID).Scan(
			&progress.DaysPracticed,
			&progress.TotalMinutes,
			&progress.MistakesFixed,
			&avgAccuracy,
		)

		if err != nil {
			log.Printf("⚠️ [WeeklyReport] DB डेटा फेच एरर (UID: %s): %v", studentUID, err)
		} else {
			progress.AccuracyScore = int(avgAccuracy)
		}
	}

	// बोली और प्रदर्शन के आधार पर प्रोत्साहन संदेश
	progress.EncouragementMsg = buildEncouragement(name, progress.DaysPracticed, dialect)
	return progress
}

func buildEncouragement(name string, days int, dialect string) string {
	d := strings.ToUpper(strings.TrimSpace(dialect))

	if days >= 5 {
		switch d {
		case "BAGRI", "MARWARI":
			return fmt.Sprintf("शाबाश %s! इस हफ्ते %d दिन घणो आछो अभ्यास करयो। मेडल पक्को!", name, days)
		case "HARYANVI":
			return fmt.Sprintf("घणा बढ़िया %s लाडले! पूरे %d दिन लगातार लठ गाड़ दिया!", name, days)
		case "BHOJPURI":
			return fmt.Sprintf("बहुत शानदार %s बबुआ! %d दिन लगातार पढ़ के गर्दा उड़ा दिए!", name, days)
		case "PUNJABI":
			return fmt.Sprintf("ਬਹੁਤ ਵਧੀਆ %s ਪੁੱਤਰ! %d ਦਿਨ ਲਗਾਤਾਰ ਮਿਹਨਤ ਕਰਕੇ ਚੱਕ ਤੇ ਫੱਟੇ!", name, days)
		default:
			return fmt.Sprintf("शानदार %s! इस हफ्ते %d दिन निरंतर 15-मिनट अभ्यास पूरा किया।", name, days)
		}
	}

	return fmt.Sprintf("अच्छा प्रयास %s! अगले हफ्ते रोज 15-मिनट अभ्यास करके 7-दिन का मेडल जीतें।", name)
}

func (w *WeeklyReportService) FormatWhatsAppReportCard(wp WeeklyProgress) string {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	today := time.Now().In(loc).Format("02 Jan 2006")

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

