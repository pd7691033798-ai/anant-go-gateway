package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

type SecuritySuite struct {
	secretKey  []byte
	db         *sql.DB
	rateLimits map[string]*Tracker
	mu         sync.Mutex
}

type Tracker struct {
	LastRequest time.Time
	DailyScans  int
}

func NewSecuritySuite(secretKey string, db *sql.DB) *SecuritySuite {
	return &SecuritySuite{
		secretKey:  []byte(secretKey),
		db:         db,
		rateLimits: make(map[string]*Tracker),
	}
}

func (s *SecuritySuite) ValidateRateLimit(phone string, maxDaily int) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	t, exists := s.rateLimits[phone]
	if !exists {
		s.rateLimits[phone] = &Tracker{LastRequest: now, DailyScans: 1}
		return true, ""
	}

	if now.Sub(t.LastRequest) < 30*time.Second {
		return false, "कृपया अगले स्कैन के लिए 30 सेकंड प्रतीक्षा करें।"
	}
	if t.DailyScans >= maxDaily {
		return false, "दैनिक स्कैन सीमा समाप्त हो गई है।"
	}

	t.LastRequest = now
	t.DailyScans++
	return true, ""
}

func (s *SecuritySuite) GenerateHash(data []byte) string {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *SecuritySuite) VerifySubmission(phone, ocrText, imgHash string) (bool, string, string) {
	if len(strings.TrimSpace(ocrText)) < 12 {
		return false, "BLANK_PAGE", "पन्ने पर लिखावट स्पष्ट नहीं आ पाई है। कृपया अच्छी रोशनी में दोबारा फोटो भेजें!"
	}

	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM submission_logs WHERE student_phone = $1 AND image_hash = $2`, phone, imgHash).Scan(&count)
	if count > 0 {
		return false, "DUPLICATE_PAGE", "यह पन्ना पहले जाँचा जा चुका है। नया पन्ना स्कैन करें!"
	}

	s.db.Exec(`INSERT INTO submission_logs (student_phone, image_hash) VALUES ($1, $2)`, phone, imgHash)
	s.db.Exec(`UPDATE users SET last_scan_date = CURRENT_DATE, streak_count = streak_count + 1 WHERE phone = $1`, phone)

	return true, "APPROVED", "सबमिशन स्वीकृत"
}
