package security

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type DynamicSessionWatermark struct {
	SessionCode string
	StudentPhone string
	ValidUntil  time.Time
}

// GenerateSessionWatermark: कॉपी के शीर्ष पर हाथ से लिखने हेतु 3-अंकों का यूनिक कोड
func GenerateSessionWatermark(phone string) (string, time.Time, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900))
	if err != nil {
		return "", time.Time{}, err
	}
	code := fmt.Sprintf("%03d", n.Int64()+100)
	expiresAt := time.Now().Add(18 * time.Minute) // 15 मिनट टेस्ट + 3 मिनट अपलोड बफर
	return code, expiresAt, nil
}
