package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type ProductionVaultManager struct {
	secretKey []byte
}

func NewProductionVaultManager(key string) *ProductionVaultManager {
	return &ProductionVaultManager{secretKey: []byte(key)}
}

// vCard 1-टैप कॉन्टैक्ट सेवर हैंडलर (9664006651)
func (v *ProductionVaultManager) ServeVCardHandler(w http.ResponseWriter, r *http.Request) {
	vcardContent := strings.Join([]string{
		"BEGIN:VCARD",
		"VERSION:3.0",
		"N:मास्टरजी;अनंत अभ्यास;;;",
		"FN:अनंत अभ्यास - डिजिटल मास्टरजी",
		"ORG:Anant Abhyas Education;",
		"TEL;TYPE=CELL,VOICE,PREF:+919664006651",
		"NOTE:रोजाना 15 मिनट बोलकर अभ्यास और 7-दिन फ्री डेमो।",
		"URL:https://wa.me/919664006651?text=राम%20राम%20सा%20मुझे%20फ्री%20डेमो%20चाहिए",
		"END:VCARD",
	}, "\r\n")

	w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"Anant_Abhyas.vcf\"")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(vcardContent))
}

// AES-256 डिक्रिप्शन वॉल्ट
func (v *ProductionVaultManager) DecryptBusinessLogic(ciphertext, nonce []byte) (string, error) {
	block, err := aes.NewCipher(v.secretKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("डिक्रिप्शन विफल: सुरक्षा छेड़छाड़")
	}
	return string(plaintext), nil
}

// एंटी-डिबगिंग व एग्जीक्यूशन इंटीग्रिटी शील्ड
func (v *ProductionVaultManager) VerifyExecutionIntegrity() {
	if _, err := os.Stat("/proc/self/status"); err == nil {
		data, _ := os.ReadFile("/proc/self/status")
		if !strings.Contains(string(data), "TracerPid:\t0") {
			log.Fatal("🚨 सुरक्षा उल्लंघन: डिबगर पकड़ा गया! सिस्टम बंद हो रहा है।")
		}
	}
}
