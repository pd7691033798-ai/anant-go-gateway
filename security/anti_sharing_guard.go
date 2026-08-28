package security

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type SharingVerdict struct {
	IsBlocked   bool   `json:"is_blocked"`
	ReasonCode  string `json:"reason_code"`
	UserMessage string `json:"user_message"`
}

type AntiSharingGuard struct {
	db *sql.DB
}

func NewAntiSharingGuard(db *sql.DB) *AntiSharingGuard {
	return &AntiSharingGuard{db: db}
}

// EvaluateSharingRisk ग्रेड विचलन, लिखावट समानता और दैनिक कोटा की जाँच करता है
func (a *AntiSharingGuard) EvaluateSharingRisk(ctx context.Context, phone string, registeredGrade, scannedGrade int, similarity float64, dailyUsed, maxAllowed int) *SharingVerdict {
	cleanPhone := strings.TrimSpace(phone)

	// 1. दैनिक कोटा समाप्ति जाँच
	if dailyUsed >= maxAllowed {
		return &SharingVerdict{
			IsBlocked:   true,
			ReasonCode:  "QUOTA_EXHAUSTED",
			UserMessage: "आज का दैनिक अभ्यास कोटा पूर्ण हो चुका है। कल पुनः अभ्यास करें।",
		}
	}

	dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 2. कक्षा/ग्रेड मिसमैच डिटेक्शन (±1 कक्षा से अधिक अंतर)
	gradeDiff := scannedGrade - registeredGrade
	if gradeDiff > 1 || gradeDiff < -1 {
		if a.db != nil && cleanPhone != "" {
			_, _ = a.db.ExecContext(dbCtx, `
				UPDATE users 
				SET sharing_suspicion_score = sharing_suspicion_score + 30,
				    detected_grade_drift_count = detected_grade_drift_count + 1 
				WHERE phone = $1
			`, cleanPhone)

			_, _ = a.db.ExecContext(dbCtx, `
				INSERT INTO multi_user_violations (student_phone, detected_issue, confidence_score) 
				VALUES ($1, $2, $3)
			`, cleanPhone, "GRADE_DRIFT_MISMATCH", 0.90)
		}

		return &SharingVerdict{
			IsBlocked:   true,
			ReasonCode:  "GRADE_MISMATCH",
			UserMessage: fmt.Sprintf("यह पन्ना कक्षा %d के पाठ्यक्रम से भिन्न लग रहा है। कृपया अपनी निर्धारित कक्षा का कार्य ही स्कैन करें।", registeredGrade),
		}
	}

	// 3. हैंडराइटिंग और लिखावट समानता जाँच (थ्रेशोल्ड < 0.55)
	if similarity < 0.55 {
		if a.db != nil && cleanPhone != "" {
			_, _ = a.db.ExecContext(dbCtx, `
				UPDATE users 
				SET sharing_suspicion_score = sharing_suspicion_score + 25 
				WHERE phone = $1
			`, cleanPhone)

			_, _ = a.db.ExecContext(dbCtx, `
				INSERT INTO multi_user_violations (student_phone, detected_issue, confidence_score) 
				VALUES ($1, $2, $3)
			`, cleanPhone, "WRITING_STYLE_MISMATCH", 1.0-similarity)
		}

		return &SharingVerdict{
			IsBlocked:   true,
			ReasonCode:  "WRITER_MISMATCH",
			UserMessage: "आज की लिखावट आपकी नियमित लिखावट शैली से भिन्न प्रतीत हो रही है। कृपया अपनी स्वाभाविक लिखावट में ही गृहकार्य भेजें।",
		}
	}

	return &SharingVerdict{
		IsBlocked:   false,
		ReasonCode:  "AUTHENTIC",
		UserMessage: "सत्यापित",
	}
}

