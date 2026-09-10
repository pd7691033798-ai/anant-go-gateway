package offline

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateSMSActivationToken(deviceUUID string, grade int, secretKey string) string {
	data := fmt.Sprintf("%s:%d", deviceUUID, grade)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(data))
	sig := hex.EncodeToString(mac.Sum(nil))
	return sig[:6] // 6-अंकीय 2G SMS कोड
}
