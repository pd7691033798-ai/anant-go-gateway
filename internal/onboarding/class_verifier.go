package onboarding

import "database/sql"

type ClassVerificationEngine struct {
	db *sql.DB
}

func NewClassVerificationEngine(db *sql.DB) *ClassVerificationEngine {
	return &ClassVerificationEngine{db: db}
}

func (c *ClassVerificationEngine) SaveVerifiedClassIntent(studentUID string, baseClass int, intent string) {
	_, _ = c.db.Exec(`UPDATE students SET current_school_class = $1, learning_intent = $2 WHERE uid = $3`, baseClass, intent, studentUID)
}
