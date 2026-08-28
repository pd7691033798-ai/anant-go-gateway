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
	"strings"
)

// CryptoVault संवेदनशील डेटा के फ़ील्ड-लेवल एन्क्रिप्शन और हैशिंग का प्रबंधन करता है
type CryptoVault struct {
	secretKey []byte // AES-256 के लिए 32-बाइट की सुरक्षित मास्टर की
}

// NewCryptoVault नया CryptoVault इंस्टेंस इनिशियलाइज़ करता है
func NewCryptoVault(masterKeyHex string) (*CryptoVault, error) {
	trimmed := strings.TrimSpace(masterKeyHex)
	if trimmed == "" {
		return nil, errors.New("मास्टर सीक्रेट की खाली नहीं हो सकती")
	}

	key, err := hex.DecodeString(trimmed)
	if err != nil || len(key) != 32 {
		// यदि 64-हेक्स वर्ण नहीं हैं, तो SHA-256 के ज़रिए 32-बाइट की उत्पन्न करें
		h := sha256.Sum256([]byte(trimmed))
		return &CryptoVault{secretKey: h[:]}, nil
	}

	return &CryptoVault{secretKey: key}, nil
}

// 1. EncryptSensitiveField: AES-256-GCM फ़ील्ड-लेवल एन्क्रिप्शन
func (v *CryptoVault) EncryptSensitiveField(plainText string) (string, error) {
	if plainText == "" {
		return "", nil
	}

	block, err := aes.NewCipher(v.secretKey)
	if err != nil {
		return "", fmt.Errorf("साइफ़र निर्माण में त्रुटि: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM मोड निर्माण में त्रुटि: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("सुरक्षित नॉनस जनरेशन में त्रुटि: %w", err)
	}

	// नॉनस को साइफ़रटेक्स्ट के साथ सील किया जाता है
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

// 2. DecryptSensitiveField: AES-256-GCM फ़ील्ड-लेवल डिक्रिप्शन
func (v *CryptoVault) DecryptSensitiveField(cipherHex string) (string, error) {
	if strings.TrimSpace(cipherHex) == "" {
		return "", nil
	}

	data, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", fmt.Errorf("अमान्य हेक्स डेटा: %w", err)
	}

	block, err := aes.NewCipher(v.secretKey)
	if err != nil {
		return "", fmt.Errorf("साइफ़र निर्माण में त्रुटि: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM मोड निर्माण में त्रुटि: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("अमान्य साइफ़रटेक्स्ट लंबाई: डेटा अधूरा या दूषित है")
	}

	nonce, actualCipher := data[:nonceSize], data[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, actualCipher, nil)
	if err != nil {
		return "", errors.New("डिक्रिप्शन असफल: डेटा से छेड़छाड़ की गई है या अनधिकृत की है")
	}

	return string(plainText), nil
}

// 3. SignPayload: HMAC-SHA256 सिग्नेचर जनरेटर (वेबहुक्स, एक्सपायरिंग लिंक्स और ऑथ टोकन्स के लिए)
func (v *CryptoVault) SignPayload(payload string) string {
	h := hmac.New(sha256.New, v.secretKey)
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// 4. VerifySignature: टाइमिंग अटैक सुरक्षित HMAC-SHA256 सत्यापन
func (v *CryptoVault) VerifySignature(payload, signatureHex string) bool {
	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}

	h := hmac.New(sha256.New, v.secretKey)
	h.Write([]byte(payload))
	expectedSig := h.Sum(nil)

	return hmac.Equal(sigBytes, expectedSig)
}

// 5. HashIdentifier: डिवाइस फ़िंगरप्रिंटिंग और सबमिशन हैश (One-way Deterministic SHA-256)
func (v *CryptoVault) HashIdentifier(rawID string) string {
	if rawID == "" {
		return ""
	}
	h := sha256.Sum256([]byte(rawID))
	return hex.EncodeToString(h[:])
}
