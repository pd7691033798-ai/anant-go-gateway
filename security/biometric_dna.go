package security

import "database/sql"

type BiometricDNAService struct {
	db *sql.DB
}

func NewBiometricDNAService(db *sql.DB) *BiometricDNAService {
	return &BiometricDNAService{db: db}
}

func (b *BiometricDNAService) VerifyHandwritingDNA(phone string, similarity float64) (bool, string) {
	if similarity < 0.60 {
		return false, "आज की लिखावट आपकी व्यक्तिगत प्रोफ़ाइल से भिन्न लग रही है। कृपया अपनी स्वाभाविक लिखावट में ही 1 पेज भेजें।"
	}
	return true, "सत्यापित लिखावट"
}

