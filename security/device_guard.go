package security

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DeviceAction string

const (
	ActionAllowAccess    DeviceAction = "ALLOW_ACCESS"     // अधिकृत डिवाइस - सामान्य अभ्यास चालू रखें
	ActionShowOnboarding DeviceAction = "SHOW_ONBOARDING"  // ऐप शेयर किया गया है - नए यूज़र को डेमो ऑनबोर्डिंग दिखाएँ
	ActionForceLogout    DeviceAction = "FORCE_LOGOUT"     // किसी अन्य डिवाइस पर लॉगिन हुआ - पुराना सेशन बंद करें
)

type DeviceGuardService struct {
	db        *sql.DB
	jwtSecret string
}

func NewDeviceGuardService(db *sql.DB, jwtSecret string) *DeviceGuardService {
	if jwtSecret == "" {
		jwtSecret = "anant-abhyas-default-secure-key-2026"
	}
	return &DeviceGuardService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// HandleAppLaunch जब ऐप खोला जाता है, यह तय करता है कि पुराना डेटा दिखाना है या नया डेमो
func (s *DeviceGuardService) HandleAppLaunch(ctx context.Context, studentUID, incomingDeviceID string) (DeviceAction, string, error) {
	if strings.TrimSpace(studentUID) == "" || strings.TrimSpace(incomingDeviceID) == "" {
		// कोई लॉगिन टोकन नहीं है -> नया यूज़र
		return ActionShowOnboarding, "नए विद्यार्थी का स्वागत है! 7-दिन का फ़्री डेमो शुरू करें।", nil
	}

	if s.db == nil {
		return ActionAllowAccess, "", nil
	}

	dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var activeDeviceID, sessionToken string
	var isActive bool

	query := `SELECT active_device_id, session_token, is_active FROM users WHERE student_uid = $1 OR phone = $1`
	err := s.db.QueryRowContext(dbCtx, query, studentUID).Scan(&activeDeviceID, &sessionToken, &isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// यूज़र नहीं मिला -> ऐप शेयर हुआ है, नया ऑनबोर्डिंग दें
			return ActionShowOnboarding, "अनंत अभ्यास में आपका स्वागत है! अपना फ़ोन नंबर दर्ज करें।", nil
		}
		return ActionAllowAccess, "", err
	}

	if !isActive {
		return ActionShowOnboarding, "आपका खाता निष्क्रिय है। कृपया नवीनीकरण करें।", nil
	}

	// यदि डिवाइस आईडी मेल नहीं खाता (ऐप शेयर किया गया या अन्य फोन में खोला गया)
	if activeDeviceID != "" && activeDeviceID != incomingDeviceID {
		// पुराने यूज़र का संवेदनशील डेटा लीक न हो, नए यूज़र को फ़्रेश स्क्रीन दिखाएँ
		return ActionShowOnboarding, "अनंत अभ्यास में आपका स्वागत है! 7-दिन का फ़्री डेमो शुरू करें।", nil
	}

	return ActionAllowAccess, "प्रमाणीकरण सफल", nil
}

// BindDeviceToSession लॉगिन या सफल डेमो एक्टिवेशन पर डिवाइस को बाँधना
func (s *DeviceGuardService) BindDeviceToSession(ctx context.Context, studentUID, deviceID, newSessionToken string) error {
	if s.db == nil || strings.TrimSpace(studentUID) == "" {
		return fmt.Errorf("डेटाबेस या यूज़र आईडी अमान्य है")
	}

	dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		UPDATE users 
		SET active_device_id = $1,
		    session_token = $2,
		    last_login_at = NOW(),
		    updated_at = NOW()
		WHERE student_uid = $3 OR phone = $3`

	_, err := s.db.ExecContext(dbCtx, query, deviceID, newSessionToken, studentUID)
	return err
}

// GenerateExpiringWebLink Chrome/ब्राउज़र में अभ्यास खोलने के लिए 60 मिनट की मियाद वाला साइन्ड लिंक बनाता है
func (s *DeviceGuardService) GenerateExpiringWebLink(studentUID string, validityMinutes int) string {
	if validityMinutes <= 0 {
		validityMinutes = 60
	}

	expiryUnix := time.Now().Add(time.Duration(validityMinutes) * time.Minute).Unix()
	payload := fmt.Sprintf("%s:%d", studentUID, expiryUnix)

	h := hmac.New(sha256.New, []byte(s.jwtSecret))
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("https://app.anantabhyas.com/practice?uid=%s&exp=%d&sig=%s", studentUID, expiryUnix, signature)
}

// VerifyExpiringWebLink ब्राउज़र लिंक की प्रामाणिकता और समय-सीमा की जाँच करता है
func (s *DeviceGuardService) VerifyExpiringWebLink(studentUID, expStr, signature string) (bool, string) {
	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return false, "अमान्य लिंक संरचना"
	}

	// समय-सीमा जाँच
	if time.Now().Unix() > expUnix {
		return false, "अभ्यास लिंक की समय-सीमा (60 मिनट) समाप्त हो चुकी है। कृपया WhatsApp से नया लिंक मंगाएँ।"
	}

	// डिजिटल सिग्नेचर सत्यापन (ताकि कोई UID या समय न बदल सके)
	payload := fmt.Sprintf("%s:%d", studentUID, expUnix)
	h := hmac.New(sha256.New, []byte(s.jwtSecret))
	h.Write([]byte(payload))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return false, "अमान्य या छेड़छाड़ किया गया सुरक्षा टोकन।"
	}

	return true, "सत्यापन सफल"
}
