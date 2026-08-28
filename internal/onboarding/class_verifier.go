package onboarding

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

type ClassVerificationEngine struct {
	db *sql.DB
}

func NewClassVerificationEngine(db *sql.DB) *ClassVerificationEngine {
	return &ClassVerificationEngine{db: db}
}

// SaveVerifiedClassIntent छात्र की जांची गई कक्षा और लर्निंग इंटेंट को सुरक्षित अपडेट करता है
func (c *ClassVerificationEngine) SaveVerifiedClassIntent(studentUID string, baseClass int, intent string) error {
	studentUID = strings.TrimSpace(studentUID)
	intent = strings.TrimSpace(intent)

	if studentUID == "" {
		return errors.New("studentUID खाली नहीं हो सकता")
	}

	// कक्षा 1 से 12 के बीच ही होनी चाहिए
	if baseClass < 1 || baseClass > 12 {
		return fmt.Errorf("अमान्य कक्षा: %d (केवल कक्षा 1 से 12 मान्य है)", baseClass)
	}

	query := `
		UPDATE students 
		SET current_school_class = $1, 
		    learning_intent = $2,
		    updated_at = NOW() 
		WHERE uid = $3`

	result, err := c.db.Exec(query, baseClass, intent, studentUID)
	if err != nil {
		log.Printf("❌ [ClassVerifier] DB अपडेट त्रुटि (UID: %s): %v", studentUID, err)
		return fmt.Errorf("डेटाबेस अपडेट विफल: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("प्रभावित पंक्तियों की जांच विफल: %w", err)
	}

	if rowsAffected == 0 {
		log.Printf("⚠️ [ClassVerifier] छात्र UID नहीं मिला: %s", studentUID)
		return errors.New("छात्र रिकॉर्ड नहीं मिला")
	}

	log.Printf("✅ [ClassVerifier] छात्र (UID: %s) की कक्षा %d व इंटेंट '%s' सफलतापूर्वक अपडेट किया गया।", studentUID, baseClass, intent)
	return nil
}
