package security

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type HandwritingVerdict struct {
	IsValid     bool    `json:"is_valid"`
	Score       float64 `json:"score"`
	UserMessage string  `json:"user_message"`
}

type BiometricDNAService struct {
	db *sql.DB
}

func NewBiometricDNAService(db *sql.DB) *BiometricDNAService {
	return &BiometricDNAService{db: db}
}

// VerifyHandwritingDNA छात्र की हस्तलिपि डीएनए (Handwriting Vector) की समानता जाँचता है
func (b *BiometricDNAService) VerifyHandwritingDNA(ctx context.Context, phone string, similarity float64) *HandwritingVerdict {
	cleanPhone := strings.TrimSpace(phone)

	// थ्रेशोल्ड 0.60 (60% समानता आवश्यक)
	if similarity < 0.60 {
		if b.db != nil && cleanPhone != "" {
			dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			// 1. यूज़र का सस्पिशन स्कोर 20 बढ़ाएँ
			_, _ = b.db.ExecContext(dbCtx, `
				UPDATE users 
				SET sharing_suspicion_score = sharing_suspicion_score + 20 
				WHERE phone = $1
			`, cleanPhone)

			// 2. वॉयलेशन लॉग में हैंडराइटिंग विसंगति दर्ज करें
			_, _ = b.db.ExecContext(dbCtx, `
				INSERT INTO multi_user_violations (student_phone, detected_issue, confidence_score) 
				VALUES ($1, $2, $3)
			`, cleanPhone, "BIOMETRIC_DNA_HANDWRITING_MISMATCH", 1.0-similarity)
		}

		return &HandwritingVerdict{
			IsValid:     false,
			Score:       similarity,
			UserMessage: "आज की लिखावट आपकी पंजीकृत प्रोफ़ाइल से भिन्न लग रही है। कृपया अपनी स्वाभाविक लिखावट में ही गृहकार्य भेजें।",
		}
	}

	return &HandwritingVerdict{
		IsValid:     true,
		Score:       similarity,
		UserMessage: "सत्यापित लिखावट (Handwriting DNA Verified)",
	}
}
