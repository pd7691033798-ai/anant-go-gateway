package tier2

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type AsyncEngineSuite struct {
	db *sql.DB
}

func NewAsyncEngineSuite(db *sql.DB) *AsyncEngineSuite {
	return &AsyncEngineSuite{db: db}
}

// AsyncEvaluateCheating बहुत तेज़ उत्तर देने (प्रति प्रश्न 2 सेकंड से कम) पर सत्र को बैकग्राउंड में फ्लैग करता है
func (a *AsyncEngineSuite) AsyncEvaluateCheating(sessionID, studentUID string, timeTakenSeconds, totalQuestions int) {
	sessionID = strings.TrimSpace(sessionID)
	studentUID = strings.TrimSpace(studentUID)

	if sessionID == "" || totalQuestions <= 0 {
		return
	}

	// यदि प्रति प्रश्न औसतन 2 सेकंड से भी कम समय लगा है (अस्वाभाविक गति)
	if timeTakenSeconds < (totalQuestions * 2) {
		go func(sID, uid string, timeTaken, qCount int) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			query := `
				UPDATE practice_sessions 
				SET is_flagged = TRUE, 
				    flag_reason = 'RAPID_SUBMISSION_ANOMALY',
				    updated_at = NOW() 
				WHERE id = $1`

			result, err := a.db.ExecContext(ctx, query, sID)
			if err != nil {
				log.Printf("❌ [CheatingDetector] फ्लैग अपडेट त्रुटि (Session: %s, Student: %s): %v", sID, uid, err)
				return
			}

			rows, _ := result.RowsAffected()
			if rows > 0 {
				log.Printf("⚠️ [चीटिंग चेतावनी उपकरण] छात्र %s (सत्र: %s) को फ्लैग किया गया। कुल प्रश्न: %d, लगा समय: %d सेकंड",
					uid, sID, qCount, timeTaken)
			}
		}(sessionID, studentUID, timeTakenSeconds, totalQuestions)
	}
}

// GetActiveStudentCount कुल सक्रिय छात्रों की वास्तविक संख्या लौटाता है (मल्टी-चाइल्ड सहित)
func (a *AsyncEngineSuite) GetActiveStudentCount() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var totalStudents int
	// मल्टी-चाइल्ड परिवार होने पर कुल बच्चों (total_children) का योग, कम से कम 1 प्रति एक्टिव यूजर
	query := `
		SELECT COALESCE(SUM(GREATEST(total_children, 1)), 0) 
		FROM users 
		WHERE is_active = TRUE`

	err := a.db.QueryRowContext(ctx, query).Scan(&totalStudents)
	if err != nil {
		log.Printf("❌ [AsyncEngineSuite] सक्रिय छात्र संख्या गणना विफल: %v", err)
		return 0, fmt.Errorf("सक्रिय छात्र संख्या प्राप्त करने में त्रुटि: %w", err)
	}

	return totalStudents, nil
}
