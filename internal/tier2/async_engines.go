package tier2

import (
	"database/sql"
	"log"
)

type AsyncEngineSuite struct {
	db *sql.DB
}

func NewAsyncEngineSuite(db *sql.DB) *AsyncEngineSuite {
	return &AsyncEngineSuite{db: db}
}

func (a *AsyncEngineSuite) AsyncEvaluateCheating(sessionID, studentUID string, timeTakenSeconds, totalQuestions int) {
	go func() {
		if totalQuestions > 0 && timeTakenSeconds < (totalQuestions*2) {
			log.Printf("⚠️ [चीटिंग चेतावनी उपकरण] छात्र %s (सत्र: %s)", studentUID, sessionID)
			_, _ = a.db.Exec(`UPDATE practice_sessions SET is_flagged = TRUE WHERE id = $1`, sessionID)
		}
	}()
}

func (a *AsyncEngineSuite) GetActiveStudentCount() (int, error) {
	var count int
	err := a.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_active = TRUE`).Scan(&count)
	return count, err
}
