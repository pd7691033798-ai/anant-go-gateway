package student

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type SundayProactiveTrigger struct {
	db *sql.DB
}

func NewSundayProactiveTrigger(db *sql.DB) *SundayProactiveTrigger {
	return &SundayProactiveTrigger{db: db}
}

// isSundayInIndia जांचता है कि भारत (IST) में आज रविवार है या नहीं
func isSundayInIndia() bool {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// फॉलबैक: UTC + 5:30
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	return time.Now().In(loc).Weekday() == time.Sunday
}

func (s *SundayProactiveTrigger) Get13thMinuteSundayPrompt(currentClass int) (bool, string) {
	if !isSundayInIndia() {
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
	if !isSundayInIndia() {
		return false, "यह सुविधा केवल रविवार को उपलब्ध है।", 0
	}

	studentUID = strings.TrimSpace(studentUID)
	input := strings.ToLower(strings.TrimSpace(inputKey))

	// '9', 'हाँ', 'ha', 'haan', 'yes', 'y' को स्वीकार करना
	isConsent := input == "9" || input == "हाँ" || input == "ha" || input == "haan" || input == "yes" || input == "y"

	if isConsent && studentUID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		query := `
			INSERT INTO weekly_advance_metrics (student_uid, week_start_date, sunday_advance_seconds_used, is_consent_given, updated_at)
			VALUES ($1, DATE_TRUNC('week', CURRENT_DATE), 0, TRUE, NOW())
			ON CONFLICT (student_uid, week_start_date)
			DO UPDATE SET is_consent_given = TRUE, updated_at = NOW()`

		if _, err := s.db.ExecContext(ctx, query, studentUID); err != nil {
			log.Printf("❌ [SundayTrigger] DB अपडेट त्रुटि (UID: %s): %v", studentUID, err)
		} else {
			log.Printf("🌟 [संडे 10-मिनट एक्सटेंशन स्वीकृत] छात्र %s ने अग्रिम स्लॉट अनलॉक किया।", studentUID)
		}

		return true, "शाबाश बेटा! आपका 10 मिनट का अग्रिम ओरिएंटेशन शुरू हो रहा है।", 600
	}

	return false, "सत्र सामान्य रूप से समाप्त हो रहा है।", 0
}
