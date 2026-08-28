package security

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

type SubmissionVerdict struct {
	IsApproved  bool   `json:"is_approved"`
	ReasonCode  string `json:"reason_code"`
	UserMessage string `json:"user_message"`
}

type Tracker struct {
	LastRequest time.Time
	DailyScans  int
	TrackedDate string
}

type SecuritySuite struct {
	secretKey  []byte
	db         *sql.DB
	rateLimits map[string]*Tracker
	mu         sync.Mutex
}

func NewSecuritySuite(secretKey string, db *sql.DB) *SecuritySuite {
	cleanKey := strings.TrimSpace(secretKey)
	if cleanKey == "" {
		cleanKey = "anant-abhyas-secure-master-salt-2026"
	}

	return &SecuritySuite{
		secretKey:  []byte(cleanKey),
		db:         db,
		rateLimits: make(map[string]*Tracker),
	}
}

// ValidateRateLimit 30-सेकंड की कूल-डाउन और दैनिक स्कैन कोटा की जाँच करता है
func (s *SecuritySuite) ValidateRateLimit(phone string, maxDaily int) (bool, string) {
	cleanPhone := strings.TrimSpace(phone)
	if cleanPhone == "" {
		return false, "अमान्य फ़ोन नंबर"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	todayStr := now.Format("2006-01-02")

	t, exists := s.rateLimits[cleanPhone]
	if !exists {
		s.rateLimits[cleanPhone] = &Tracker{
			LastRequest: now,
			DailyScans:  1,
			TrackedDate: todayStr,
		}
		return true, ""
	}

	// 1. यदि नई तारीख शुरू हो गई है, तो दैनिक काउंटर रीसेट करें
	if t.TrackedDate != todayStr {
		t.DailyScans = 1
		t.TrackedDate = todayStr
		t.LastRequest = now
		return true, ""
	}

	// 2. 30 सेकंड का कूलडाउन सुरक्षा
	if now.Sub(t.LastRequest) < 30*time.Second {
		remainingSec := int(30 - now.Sub(t.LastRequest).Seconds())
		return false, "कृपया अगले स्कैन के लिए थोड़ा इंतज़ार करें (" + string(rune(remainingSec)) + " सेकंड शेष)।"
	}

	// 3. दैनिक कोटा सीमा जाँच
	if t.DailyScans >= maxDaily {
		return false, "आज का दैनिक अभ्यास कोटा पूर्ण हो चुका है। कल पुनः अभ्यास करें!"
	}

	t.LastRequest = now
	t.DailyScans++
	return true, ""
}

// GenerateHash डेटा पेलोड के लिए HMAC-SHA256 डिजिटल सिग्नेचर बनाता है
func (s *SecuritySuite) GenerateHash(data []byte) string {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySubmission पन्ने की लिखावट, डुप्लीकेट इमेज हैश और स्ट्रीक को सुरक्षित रूप से संसाधित करता है
func (s *SecuritySuite) VerifySubmission(ctx context.Context, phone, ocrText, imgHash string) *SubmissionVerdict {
	cleanPhone := strings.TrimSpace(phone)
	cleanHash := strings.TrimSpace(imgHash)

	// 1. खाली या अत्यधिक अस्पष्ट पन्ना जांच
	if len(strings.TrimSpace(ocrText)) < 12 {
		return &SubmissionVerdict{
			IsApproved:  false,
			ReasonCode:  "BLANK_PAGE",
			UserMessage: "पन्ने पर लिखावट स्पष्ट नहीं आ पाई है। कृपया अच्छी रोशनी में दोबारा साफ़ फोटो भेजें!",
		}
	}

	if s.db == nil {
		return &SubmissionVerdict{
			IsApproved:  true,
			ReasonCode:  "APPROVED_STANDALONE",
			UserMessage: "सबमिशन स्वीकृत",
		}
	}

	dbCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// 2. डुप्लीकेट सबमिशन सुरक्षा (Duplicate Page Detection)
	var count int
	err := s.db.QueryRowContext(dbCtx, `SELECT COUNT(*) FROM submission_logs WHERE student_phone = $1 AND image_hash = $2`, cleanPhone, cleanHash).Scan(&count)
	if err == nil && count > 0 {
		return &SubmissionVerdict{
			IsApproved:  false,
			ReasonCode:  "DUPLICATE_PAGE",
			UserMessage: "यह पन्ना पहले ही जाँचा जा चुका है। कृपया गृहकार्य का नया पन्ना स्कैन करें!",
		}
	}

	// 3. नया सबमिशन लॉग दर्ज करना
	_, _ = s.db.ExecContext(dbCtx, `INSERT INTO submission_logs (student_phone, image_hash) VALUES ($1, $2)`, cleanPhone, cleanHash)

	// 4. स्ट्रीक और अंतिम स्कैन तारीख अद्यतन (एक दिन में केवल एक बार स्ट्रीक बढ़े)
	updateStreakQuery := `
		UPDATE users 
		SET streak_count = CASE 
				WHEN last_scan_date = CURRENT_DATE THEN streak_count
				WHEN last_scan_date = CURRENT_DATE - INTERVAL '1 day' THEN streak_count + 1
				ELSE 1 
		    END,
		    last_scan_date = CURRENT_DATE,
		    consecutive_missed_days = 0
		WHERE phone = $1
	`
	_, _ = s.db.ExecContext(dbCtx, updateStreakQuery, cleanPhone)

	return &SubmissionVerdict{
		IsApproved:  true,
		ReasonCode:  "APPROVED",
		UserMessage: "सबमिशन सफलतापूर्वक स्वीकृत और जाँचा गया।",
	}
}
