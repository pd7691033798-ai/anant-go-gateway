package security

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type GeoTravelVerdict struct {
	IsAllowed   bool    `json:"is_allowed"`
	ReasonCode  string  `json:"reason_code"`
	UserMessage string  `json:"user_message"`
	Confidence  float64 `json:"confidence"`
}

type GeoTravelService struct {
	db *sql.DB
}

func NewGeoTravelService(db *sql.DB) *GeoTravelService {
	return &GeoTravelService{db: db}
}

// CheckTravelEvent भौगोलिक स्थान बदलाव और डिवाइस/हैंडराइटिंग विश्वसनीयता की पुष्टि करता है
func (g *GeoTravelService) CheckTravelEvent(ctx context.Context, phone, currentCity, deviceHash string, handwritingSimilarity float64) *GeoTravelVerdict {
	cleanPhone := strings.TrimSpace(phone)
	cleanCity := strings.TrimSpace(currentCity)
	cleanDeviceHash := strings.TrimSpace(deviceHash)

	if g.db == nil || cleanPhone == "" {
		return &GeoTravelVerdict{
			IsAllowed:   true,
			ReasonCode:  "BYPASS_NO_DB",
			UserMessage: "सत्यापित",
			Confidence:  1.0,
		}
	}

	dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var primaryDevice, lastCity string
	query := `SELECT COALESCE(primary_device_hash, ''), COALESCE(current_location_city, '') FROM users WHERE phone = $1`
	err := g.db.QueryRowContext(dbCtx, query, cleanPhone).Scan(&primaryDevice, &lastCity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &GeoTravelVerdict{
				IsAllowed:   false,
				ReasonCode:  "USER_NOT_FOUND",
				UserMessage: "उपयोगकर्ता खाता नहीं मिला।",
				Confidence:  0.0,
			}
		}
		// डेटाबेस एरर होने पर पढ़ाई न रुके, एक्सेस दें
		return &GeoTravelVerdict{
			IsAllowed:   true,
			ReasonCode:  "DB_FALLBACK",
			UserMessage: "सत्यापित",
			Confidence:  0.5,
		}
	}

	// 1. यदि पहली बार डिवाइस रजिस्टर हो रहा है
	if primaryDevice == "" {
		_, _ = g.db.ExecContext(dbCtx, `UPDATE users SET primary_device_hash = $1, current_location_city = $2 WHERE phone = $3`, cleanDeviceHash, cleanCity, cleanPhone)
		return &GeoTravelVerdict{
			IsAllowed:   true,
			ReasonCode:  "PRIMARY_DEVICE_SET",
			UserMessage: "नया अधिकृत डिवाइस पंजीकृत किया गया।",
			Confidence:  1.0,
		}
	}

	// 2. यदि अधिकृत डिवाइस है और लिखावट सामान्य है (≥ 0.65)
	if primaryDevice == cleanDeviceHash && handwritingSimilarity >= 0.65 {
		if lastCity != cleanCity && cleanCity != "" {
			_, _ = g.db.ExecContext(dbCtx, `UPDATE users SET current_location_city = $1 WHERE phone = $2`, cleanCity, cleanPhone)
		}
		return &GeoTravelVerdict{
			IsAllowed:   true,
			ReasonCode:  "VERIFIED_TRAVEL",
			UserMessage: "यात्रा/स्थान अद्यतन सत्यापित।",
			Confidence:  handwritingSimilarity,
		}
	}

	// 3. यदि डिवाइस बदल गया है, तो लिखावट की कड़ी जाँच (सख्त थ्रेशोल्ड ≥ 0.70)
	if handwritingSimilarity >= 0.70 {
		return &GeoTravelVerdict{
			IsAllowed:   true,
			ReasonCode:  "GUEST_DEVICE_VERIFIED",
			UserMessage: "वैकल्पिक डिवाइस पर लिखावट सत्यापित।",
			Confidence:  handwritingSimilarity,
		}
	}

	// 4. नया डिवाइस + खराब/भिन्न लिखावट -> संभावित खाता साझाकरण (Account Sharing / Abuse)
	_, _ = g.db.ExecContext(dbCtx, `
		UPDATE users 
		SET sharing_suspicion_score = sharing_suspicion_score + 35 
		WHERE phone = $1
	`, cleanPhone)

	_, _ = g.db.ExecContext(dbCtx, `
		INSERT INTO multi_user_violations (student_phone, detected_issue, confidence_score) 
		VALUES ($1, $2, $3)
	`, cleanPhone, "UNAUTHORIZED_DEVICE_AND_WRITING_MISMATCH", 1.0-handwritingSimilarity)

	return &GeoTravelVerdict{
		IsAllowed:   false,
		ReasonCode:  "SUSPECTED_ACCOUNT_SHARING",
		UserMessage: "अपरिचित डिवाइस और लिखावट विसंगति पाई गई। कृपया अपने पंजीकृत डिवाइस और स्वाभाविक लिखावट का उपयोग करें।",
		Confidence:  handwritingSimilarity,
	}
}
