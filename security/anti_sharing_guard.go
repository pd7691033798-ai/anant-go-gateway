package security

import (
	"database/sql"
	"fmt"
)

type SharingVerdict struct {
	IsBlocked   bool
	ReasonCode  string
	UserMessage string
}

type AntiSharingGuard struct {
	db *sql.DB
}

func NewAntiSharingGuard(db *sql.DB) *AntiSharingGuard {
	return &AntiSharingGuard{db: db}
}

func (a *AntiSharingGuard) EvaluateSharingRisk(phone string, registeredGrade, scannedGrade int, similarity float64, dailyUsed, maxAllowed int) *SharingVerdict {
	if dailyUsed >= maxAllowed {
		return &SharingVerdict{IsBlocked: true, ReasonCode: "QUOTA_EXHAUSTED", UserMessage: "आज का दैनिक अभ्यास कोटा पूर्ण हो चुका है।"}
	}

	gradeDiff := scannedGrade - registeredGrade
	if gradeDiff > 1 || gradeDiff < -1 {
		a.db.Exec(`UPDATE users SET sharing_suspicion_score = sharing_suspicion_score + 30 WHERE phone = $1`, phone)
		return &SharingVerdict{IsBlocked: true, ReasonCode: "GRADE_MISMATCH", UserMessage: fmt.Sprintf("यह पन्ना कक्षा %d से भिन्न लग रहा है।", registeredGrade)}
	}

	if similarity < 0.55 {
		a.db.Exec(`UPDATE users SET sharing_suspicion_score = sharing_suspicion_score + 25 WHERE phone = $1`, phone)
		return &SharingVerdict{IsBlocked: true, ReasonCode: "WRITER_MISMATCH", UserMessage: "आज की लिखावट भिन्न है। अपनी स्वाभाविक लिखावट में भेजें।"}
	}

	return &SharingVerdict{IsBlocked: false, ReasonCode: "AUTHENTIC", UserMessage: "सत्यापित"}
}
