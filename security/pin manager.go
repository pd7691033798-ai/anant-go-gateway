package security

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type TempPasscode struct {
	Code      string
	ExpiresAt time.Time
}

type PINManager struct {
	db              *sql.DB
	mu              sync.RWMutex
	activeOTPs      map[string]TempPasscode // phone -> OTP
	bypassCodes     map[string]TempPasscode // phone -> 15-min passcode
	alertWebhookURL string
	httpClient      *http.Client
}

func NewPINManager(db *sql.DB, alertWebhookURL string) *PINManager {
	return &PINManager{
		db:              db,
		activeOTPs:      make(map[string]TempPasscode),
		bypassCodes:     make(map[string]TempPasscode),
		alertWebhookURL: alertWebhookURL,
		httpClient:      &http.Client{Timeout: 5 * time.Second},
	}
}

func getNextMidnight() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

func hashPIN(pin, salt string) string {
	hasher := sha256.New()
	hasher.Write([]byte(pin + salt))
	return hex.EncodeToString(hasher.Sum(nil))
}

func generateSalt() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// LogSecurityEvent: डेटाबेस में सुरक्षा ऑडिट दर्ज करता है
func (pm *PINManager) LogSecurityEvent(phone, eventType, details string) {
	if pm.db == nil {
		return
	}
	query := `INSERT INTO security_audit_logs (parent_phone, event_type, details) VALUES ($1, $2, $3)`
	_, _ = pm.db.Exec(query, phone, eventType, details)
}

// SendWhatsAppAlert: बैकग्राउंड में तुरंत WhatsApp अलर्ट भेजता है
func (pm *PINManager) SendWhatsAppAlert(phone, message string) {
	if pm.alertWebhookURL == "" {
		return
	}
	go func() {
		payload := map[string]string{
			"phone":   phone,
			"message": message,
		}
		data, _ := json.Marshal(payload)
		_, _ = pm.httpClient.Post(pm.alertWebhookURL, "application/json", bytes.NewBuffer(data))
	}()
}

// VerifyPINWithDailyLock: थ्रॉटलिंग, डेली लॉक और पर्सिस्टेंट सुरक्षा के साथ सत्यापन
func (pm *PINManager) VerifyPINWithDailyLock(phone, inputPIN string) (bool, string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var (
		storedHash    sql.NullString
		salt          sql.NullString
		lockedUntil   sql.NullTime
		failedAttempts int
		lastFailedAt  sql.NullTime
	)

	query := `
		SELECT pin_hash, pin_salt, pin_locked_until, failed_attempts, last_failed_at 
		FROM users WHERE phone = $1
	`
	err := pm.db.QueryRow(query, phone).Scan(&storedHash, &salt, &lockedUntil, &failedAttempts, &lastFailedAt)
	if err != nil {
		return false, "USER_NOT_FOUND", fmt.Errorf("उपयोगकर्ता नहीं मिला")
	}

	now := time.Now()

	// 1. मिडनाइट ऑटो-रीसेट व लॉक चेक
	if lockedUntil.Valid && now.Before(lockedUntil.Time) {
		remaining := lockedUntil.Time.Sub(now).Round(time.Minute)
		return false, "DAILY_LOCK_ACTIVE", fmt.Errorf("आज के लिए पिन पूरी तरह लॉक है। कल सुबह 00:00 बजे खुलेगा (शेष समय: %v)", remaining)
	} else if lockedUntil.Valid && now.After(lockedUntil.Time) {
		// मध्यरात्रि पार होने पर डेटाबेस में लॉक रीसेट
		_, _ = pm.db.Exec(`UPDATE users SET pin_locked_until = NULL, failed_attempts = 0 WHERE phone = $1`, phone)
		failedAttempts = 0
	}

	// 2. 30 सेकंड का कूलडाउन डिले (थ्रॉटलिंग)
	if failedAttempts > 0 && lastFailedAt.Valid && now.Sub(lastFailedAt.Time) < 30*time.Second {
		waitSec := int((30 * time.Second - now.Sub(lastFailedAt.Time)).Seconds())
		return false, "THROTTLED", fmt.Errorf("कृपया %d सेकंड प्रतीक्षा करें और पुनः प्रयास करें", waitSec)
	}

	// 3. हैश मिलान
	inputHash := hashPIN(inputPIN, salt.String)
	if !storedHash.Valid || storedHash.String != inputHash {
		failedAttempts++

		// 3 गलत प्रयासों पर मिडनाइट लॉक सक्रिय
		if failedAttempts >= 3 {
			nextMid := getNextMidnight()
			_, _ = pm.db.Exec(`
				UPDATE users 
				SET pin_locked_until = $1, failed_attempts = 0, last_failed_at = $2 
				WHERE phone = $3`, nextMid, now, phone)

			pm.LogSecurityEvent(phone, "DAILY_LOCK_TRIGGERED", "3 गलत प्रयासों के कारण खाता मध्यरात्रि तक लॉक हुआ")
			pm.SendWhatsAppAlert(phone, "⚠️ सुरक्षा चेतावनी: लगातार 3 गलत पिन दर्ज किए गए हैं। सिस्टम आज रात 12 बजे तक के लिए लॉक कर दिया गया है।")

			return false, "TRIGGER_DAILY_LOCK", fmt.Errorf("3 बार गलत पिन! सुरक्षा कारणों से सिस्टम आज रात 12 बजे तक लॉक कर दिया गया है")
		}

		_, _ = pm.db.Exec(`UPDATE users SET failed_attempts = $1, last_failed_at = $2 WHERE phone = $3`, failedAttempts, now, phone)
		pm.LogSecurityEvent(phone, "PIN_FAILED", fmt.Sprintf("गलत प्रयास संख्या: %d", failedAttempts))

		return false, "INVALID_PIN", fmt.Errorf("गलत पिन। शेष प्रयास: %d", 3-failedAttempts)
	}

	// सफलता मिलने पर रीसेट
	_, _ = pm.db.Exec(`UPDATE users SET failed_attempts = 0, pin_locked_until = NULL WHERE phone = $1`, phone)
	pm.LogSecurityEvent(phone, "PIN_SUCCESS", "मास्टर पिन सत्यापन सफल")
	return true, "SUCCESS", nil
}

// IssueOneTimeBypass: 24 घंटे में अधिकतम 2 बार 15-मिनट पासकोड जारी करता है
func (pm *PINManager) IssueOneTimeBypass(parentPhone string) (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var bypassCount int
	var bypassResetAt sql.NullTime

	query := `SELECT daily_bypass_count, bypass_reset_at FROM users WHERE phone = $1`
	err := pm.db.QueryRow(query, parentPhone).Scan(&bypassCount, &bypassResetAt)
	if err != nil {
		return "", fmt.Errorf("यूजर विवरण प्राप्त करने में विफलता")
	}

	now := time.Now()
	if !bypassResetAt.Valid || now.After(bypassResetAt.Time) {
		bypassCount = 0
		bypassResetAt.Time = now.Add(24 * time.Hour)
		bypassResetAt.Valid = true
	}

	if bypassCount >= 2 {
		return "", fmt.Errorf("दैनिक सीमा समाप्त! आप 24 घंटे में अधिकतम 2 बार ही बायपास कोड ले सकते हैं")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return "", err
	}
	code := fmt.Sprintf("%04d", n.Int64()+1000)

	pm.bypassCodes[parentPhone] = TempPasscode{
		Code:      code,
		ExpiresAt: now.Add(15 * time.Minute),
	}

	bypassCount++
	_, _ = pm.db.Exec(`UPDATE users SET daily_bypass_count = $1, bypass_reset_at = $2 WHERE phone = $3`, bypassCount, bypassResetAt.Time, parentPhone)

	pm.LogSecurityEvent(parentPhone, "BYPASS_ISSUED", fmt.Sprintf("15-मिनट बायपास कोड जारी (आज का उपयोग: %d/2)", bypassCount))
	return code, nil
}

// VerifyOneTimeBypass: बच्चे द्वारा डाले गए 15-मिनट पासकोड को सत्यापित करता है
func (pm *PINManager) VerifyOneTimeBypass(parentPhone, code string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	stored, exists := pm.bypassCodes[parentPhone]
	if !exists || time.Now().After(stored.ExpiresAt) || stored.Code != code {
		return false
	}

	delete(pm.bypassCodes, parentPhone)
	pm.LogSecurityEvent(parentPhone, "BYPASS_USED", "15-मिनट बायपास कोड का उपयोग किया गया")
	return true
}

// GenerateOTP: पिन रीसेट या इमरजेंसी अनलॉक के लिए 6-अंकों का OTP
func (pm *PINManager) GenerateOTP(phone string) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	otp := fmt.Sprintf("%06d", n.Int64()+100000)

	pm.mu.Lock()
	pm.activeOTPs[phone] = TempPasscode{
		Code:      otp,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	pm.mu.Unlock()

	pm.LogSecurityEvent(phone, "OTP_GENERATED", "पिन रीसेट/अनलॉक OTP भेजा गया")
	return otp, nil
}

// UnlockViaParentEmergencyOTP: WhatsApp OTP से लॉक तत्काल हटाना
func (pm *PINManager) UnlockViaParentEmergencyOTP(phone, otp string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	stored, exists := pm.activeOTPs[phone]
	if !exists || time.Now().After(stored.ExpiresAt) || stored.Code != otp {
		return fmt.Errorf("अमान्य या समाप्त OTP")
	}

	delete(pm.activeOTPs, phone)

	_, err := pm.db.Exec(`UPDATE users SET pin_locked_until = NULL, failed_attempts = 0 WHERE phone = $1`, phone)
	if err != nil {
		return err
	}

	pm.LogSecurityEvent(phone, "EMERGENCY_UNLOCK", "अभिभावक OTP द्वारा लॉक हटाया गया")
	return nil
}

// SetOrResetMasterPIN: नया पिन हैश करके डेटाबेस में सुरक्षित करता है
func (pm *PINManager) SetOrResetMasterPIN(phone, otp, newPIN string, isInitialSetup bool) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !isInitialSetup {
		stored, exists := pm.activeOTPs[phone]
		if !exists || time.Now().After(stored.ExpiresAt) || stored.Code != otp {
			return fmt.Errorf("अमान्य या समाप्त OTP")
		}
		delete(pm.activeOTPs, phone)
	}

	if len(newPIN) != 4 {
		return fmt.Errorf("मास्टर पिन ठीक 4 अंकों का होना चाहिए")
	}

	salt, err := generateSalt()
	if err != nil {
		return fmt.Errorf("सुरक्षा साल्ट निर्माण में त्रुटि")
	}
	hash := hashPIN(newPIN, salt)

	query := `
		UPDATE users 
		SET pin_hash = $1, pin_salt = $2, pin_locked_until = NULL, failed_attempts = 0 
		WHERE phone = $3
	`
	_, err = pm.db.Exec(query, hash, salt, phone)
	if err == nil {
		pm.LogSecurityEvent(phone, "PIN_SET_SUCCESS", "मास्टर पिन सुरक्षित रूप से अपडेट हुआ")
	}
	return err
}
