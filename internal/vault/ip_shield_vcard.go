package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

type ProductionVaultManager struct {
	secretKey []byte
}

func NewProductionVaultManager(key string) (*ProductionVaultManager, error) {
	k := []byte(key)
	// AES-256 के लिए 32 बाइट्स, AES-128 के लिए 16 बाइट्स अनिवार्य हैं
	if len(k) != 16 && len(k) != 24 && len(k) != 32 {
		return nil, fmt.Errorf("अमान्य AES की लंबाई: %d बाइट्स (केवल 16, 24, या 32 बाइट्स मान्य)", len(k))
	}
	return &ProductionVaultManager{secretKey: k}, nil
}

// vCard 1-टैप कॉन्टैक्ट सेवर हैंडलर (9664006651)
func (v *ProductionVaultManager) ServeVCardHandler(w http.ResponseWriter, r *http.Request) {
	waMsg := url.QueryEscape("राम राम सा मुझे फ्री डेमो चाहिए")
	waURL := fmt.Sprintf("https://wa.me/919664006651?text=%s", waMsg)

	vcardLines := []string{
		"BEGIN:VCARD",
		"VERSION:3.0",
		"N:मास्टरजी;अनंत अभ्यास;;;",
		"FN:अनंत अभ्यास - डिजिटल मास्टरजी",
		"ORG:Anant Abhyas Education",
		"TEL;TYPE=CELL,VOICE,PREF:+919664006651",
		"NOTE:रोजाना 15 मिनट बोलकर अभ्यास और 7-दिन फ्री डेमो।",
		fmt.Sprintf("URL:%s", waURL),
		"END:VCARD",
	}

	vcardContent := strings.Join(vcardLines, "\r\n")

	w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="Anant_Abhyas.vcf"`)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(vcardContent)); err != nil {
		log.Printf("❌ [Vault] vCard डिलीवरी त्रुटि: %v", err)
	}
}

// AES-GCM डिक्रिप्शन वॉल्ट
func (v *ProductionVaultManager) DecryptBusinessLogic(ciphertext, nonce []byte) (string, error) {
	if len(v.secretKey) == 0 {
		return "", errors.New("वॉल्ट सीक्रेट की सेट नहीं है")
	}

	block, err := aes.NewCipher(v.secretKey)
	if err != nil {
		return "", fmt.Errorf("सिफर निर्माण विफल: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM ब्लॉक निर्माण विफल: %w", err)
	}

	if len(nonce) != aesGCM.NonceSize() {
		return "", fmt.Errorf("अमान्य Nonce साइज: %d (आवश्यक: %d)", len(nonce), aesGCM.NonceSize())
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("डिक्रिप्शन विफल: डेटा छेड़छाड़ या गलत सुरक्षा कुंजी")
	}

	return string(plaintext), nil
}

// एंटी-डिबगिंग व एग्जीक्यूशन इंटीग्रिटी शील्ड (Linux Production Specific)
func (v *ProductionVaultManager) VerifyExecutionIntegrity() {
	if runtime.GOOS != "linux" {
		return // नॉन-लिनक्स एनवायरनमेंट में बाईपास
	}

	statusPath := "/proc/self/status"
	if _, err := os.Stat(statusPath); err == nil {
		data, err := os.ReadFile(statusPath)
		if err != nil {
			return
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "TracerPid:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 && fields[1] != "0" {
					log.Fatal("🚨 सुरक्षा उल्लंघन: डिबगर पकड़ा गया (TracerPid != 0)! सुरक्षा कारणों से सिस्टम बंद हो रहा है।")
				}
				break
			}
		}
	}
}
