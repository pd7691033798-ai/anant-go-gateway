package student

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type SundayProactiveTrigger struct {
	db *sql.DB
}

func NewSundayProactiveTrigger(db *sql.DB) *SundayProactiveTrigger {
	return &SundayProactiveTrigger{db: db}
}

func (s *SundayProactiveTrigger) Get13thMinuteSundayPrompt(currentClass int) (bool, string) {
	if time.Now().Weekday() != time.Sunday {
		return false, ""
	}
	nextClass := currentClass + 1
	if currentClass >= 12 {
		nextClass = 12
	}
	prompt := fmt.Sprintf(
		"बेटा, आज के अभ्यास में 2 मिनट बचे हैं! आज रविवार है—अगर आप अगली कक्षा (%d) या नए हुनर की 10 मिनट की अग्रिम जानकारी चाहते हैं, तो अभी कीपैड पर 9 दबाएं या 'हाँ' बोलें।",
		nextClass,
	)
	return true, prompt
}

func (s *SundayProactiveTrigger) HandleExtensionConsent(studentUID string, currentClass int, inputKey string) (bool, string, int) {
	if time.Now().Weekday() != time.Sunday {
		return false, "यह सुविधा केवल रविवार को उपलब्ध है।", 0
	}

	if inputKey == "9" || inputKey == "हाँ" || inputKey == "ha" || inputKey == "yes" {
		log.Printf("🌟 [संडे 10-मिनट एक्सटेंशन स्वीकृत] छात्र %s ने अग्रिम स्लॉट अनलॉक किया।", studentUID)
		query := `
			INSERT INTO weekly_advance_metrics (student_uid, week_start_date, sunday_advance_seconds_used, is_consent_given)
			VALUES ($1, DATE_TRUNC('week', CURRENT_DATE), 0, TRUE)
			ON CONFLICT (student_uid, week_start_date)
			DO UPDATE SET is_consent_given = TRUE, updated_at = NOW()`
		_, _ = s.db.Exec(query, studentUID)
		return true, "शाबाश बेटा! आपका 10 मिनट का अग्रिम ओरिएंटेशन शुरू हो रहा है।", 600
	}
	return false, "सत्र सामान्य रूप से समाप्त हो रहा है।", 0
}
