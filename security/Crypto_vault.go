package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// CryptoVault संवेदनशील डेटा के फ़ील्ड-लेवल एन्क्रिप्शन और हैशिंग का प्रबंधन करता है
type CryptoVault struct {
	secretKey []byte // AES-256 के लिए 32-बाइट की सुरक्षित की (Master Secret)
}

func NewCryptoVault(masterKeyHex string) (*CryptoVault, error) {
	key, err := hex.DecodeString(masterKeyHex)
	if err != nil || len(key) != 32 {
		// यदि मास्टर की नहीं दी गई है, तो 32-बाइट डिफ़ॉल्ट (केवल प्रोडक्शन में पर्यावरण वेरिएबल से लें)
		h := sha256.Sum256([]byte(masterKeyHex))
		return &CryptoVault{secretKey: h[:]}, nil
	}
	return &CryptoVault{secretKey: key}, nil
}

// 1. AES-256-GCM फ़ील्ड-लेवल एन्क्रिप्शन (संवेदनशील टोकन्स और सीक्रेट्स के लिए)
func (v *CryptoVault) EncryptSensitiveField(plainText string) (string, error) {
	if plainText == "" {
		return "", nil
	}

	block, err := aes.NewCipher(v.secretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

// 2. AES-256-GCM फ़ील्ड-लेवल डिक्रिप्शन
func (v *CryptoVault) DecryptSensitiveField(cipherHex string) (string, error) {
	if cipherHex == "" {
		return "", nil
	}

	data, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(v.secretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("अमान्य सिफ़रटेक्स्ट लंबाई")
	}

	nonce, actualCipher := data[:nonceSize], data[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, actualCipher, nil)
	if err != nil {
		return "", errors.New("डिक्रिप्शन असफल: डेटा से छेड़छाड़ की गई है")
	}

	return string(plainText), nil
}

// 3. HMAC-SHA256 सिग्नेचर जनरेटर (वेबहुक और टाइम-बाउंड लिंक्स के लिए)
func (v *CryptoVault) SignPayload(payload string) string {
	h := hmac.New(sha256.New, v.secretKey)
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// 4. HMAC-SHA256 सिग्नेचर सत्यापन
func (v *CryptoVault) VerifySignature(payload, signatureHex string) bool {
	expectedSig := v.SignPayload(payload)
	return hmac.Equal([]byte(signatureHex), []byte(expectedSig))
}

// 5. डिवाइस फ़िंगरप्रिंटिंग हैश (One-way Hash)
func (v *CryptoVault) HashIdentifier(rawID string) string {
	h := sha256.New()
	h.Write([]byte(rawID))
	return hex.EncodeToString(h.Sum(nil))
}

