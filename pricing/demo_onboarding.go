package pricing

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type DemoStatus string

const (
	DemoActive  DemoStatus = "DEMO_ACTIVE"  // 1-7 दिन
	DemoGrace   DemoStatus = "DEMO_GRACE"   // 8-10 दिन (3 दिन का ग्रेस पीरियड)
	DemoExpired DemoStatus = "DEMO_EXPIRED" // 10 दिन पूरे
)

type DemoService struct {
	db *sql.DB
}

func NewDemoService(db *sql.DB) *DemoService {
	return &DemoService{db: db}
}

// sanitizePhone 10 अंकों का शुद्ध मोबाइल नंबर निकालता है
func sanitizePhone(phone string) string {
	clean := strings.TrimPrefix(phone, "+91")
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")
	if len(clean) >= 10 {
		return clean[len(clean)-10:]
	}
	return clean
}

// Activate7DayDemo सुरक्षित रूप से 7-दिन का फ़्री डेमो + 3-दिन ग्रेस पीरियड रिकॉर्ड सेट करता है
func (d *DemoService) Activate7DayDemo(ctx context.Context, phone, name string, grade int, state, district, dialect string) error {
	cleanPhone := sanitizePhone(phone)
	if len(cleanPhone) != 10 {
		return fmt.Errorf("अमान्य फ़ोन नंबर: %s", phone)
	}

	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		cleanName = "विद्यार्थी"
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 1. users टेबल में पेरेंट/अकाउंट स्तर पर 7 दिन डेमो + 3 दिन ग्रेस (कुल 10 दिन) दर्ज करना
	// ON CONFLICT पर केवल तभी नया टाइम सेट होगा जब पहले कभी डेमो का उपयोग न हुआ हो
	userQuery := `
		INSERT INTO users (
			phone, name, grade, state, district, preferred_dialect, 
			plan_tier, is_active, demo_days_used, plan_expires_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, 
			'DEMO', TRUE, 0, NOW() + INTERVAL '10 days', NOW(), NOW()
		)
		ON CONFLICT (phone) DO UPDATE 
		SET state = EXCLUDED.state,
		    district = EXCLUDED.district,
		    preferred_dialect = EXCLUDED.preferred_dialect,
		    updated_at = NOW()
		WHERE users.plan_tier = 'DEMO' AND users.demo_days_used < 7`

	_, err := d.db.ExecContext(dbCtx, userQuery, cleanPhone, cleanName, grade, state, district, dialect)
	if err != nil {
		return fmt.Errorf("डेमो एक्टिवेशन विफल (Users Table): %w", err)
	}

	// 2. students टेबल में छात्र प्रोफ़ाइल सुरक्षित करना
	studentUID := fmt.Sprintf("STU_%s_1", cleanPhone)
	studentQuery := `
		INSERT INTO students (
			uid, parent_phone, name, class_grade, preferred_dialect, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (uid) DO UPDATE 
		SET name = EXCLUDED.name,
		    class_grade = EXCLUDED.class_grade,
		    preferred_dialect = EXCLUDED.preferred_dialect,
		    updated_at = NOW()`

	_, err = d.db.ExecContext(dbCtx, studentQuery, studentUID, cleanPhone, cleanName, grade, dialect)
	if err != nil {
		return fmt.Errorf("छात्र रिकॉर्ड सिंक विफल (Students Table): %w", err)
	}

	return nil
}

// CheckDemoStatus 7-दिन एक्टिव डेमो, 3-दिन ग्रेस पीरियड या समाप्ति की जाँच करता है
func (d *DemoService) CheckDemoStatus(expiresAt time.Time, demoDaysUsed int) (DemoStatus, int) {
	now := time.Now()

	// यदि 10 दिन (7 दिन डेमो + 3 दिन ग्रेस) की कुल अवधि समाप्त हो चुकी है या 10 दिन पूरे हो चुके हैं
	if now.After(expiresAt) || demoDaysUsed >= 10 {
		return DemoExpired, 0
	}

	// दिन 1 से 7: एक्टिव डेमो
	if demoDaysUsed < 7 {
		remainingDays := 7 - demoDaysUsed
		return DemoActive, remainingDays
	}

	// दिन 8 से 10: 3 दिन का ग्रेस पीरियड
	remainingGrace := 10 - demoDaysUsed
	return DemoGrace, remainingGrace
}
