package featurephone

import (
	"database/sql"
	"fmt"
	"strings"
)

type FeaturePhoneEngine struct {
	db *sql.DB
}

// NewFeaturePhoneEngine डेटाबेस कनेक्शन के साथ इंजन को इनिशियलाइज करता है
func NewFeaturePhoneEngine(db *sql.DB) *FeaturePhoneEngine {
	return &FeaturePhoneEngine{db: db}
}

// GenerateIVRScriptFromDB डेटाबेस से छात्र का डेटा (नाम, राज्य, भाषा, सवाल) फेच करके IVR स्क्रिप्ट बनाता है
func (f *FeaturePhoneEngine) GenerateIVRScriptFromDB(studentID int) (string, error) {
	var studentName, stateCode, language, mathProblem string

	// डेटाबेस से असली छात्र का डेटा निकालना (भाषा कॉलम के साथ)
	query := "SELECT name, state_code, language, todays_problem FROM students WHERE id = $1"
	err := f.db.QueryRow(query, studentID).Scan(&studentName, &stateCode, &language, &mathProblem)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("student not found")
		}
		return "", err
	}

	// 1. यदि छात्र/अभिभावक की प्राथमिकता अंग्रेजी (English) है
	if strings.ToUpper(language) == "ENGLISH" {
		template := "Hello %s! Please take your notebook and pen. Today's 15-minute challenge is: %s. Press 1 after writing."
		return fmt.Sprintf(template, studentName, mathProblem), nil
	}

	// 2. 28 राज्यों और क्षेत्रीय बोलियों के स्वागत/निर्देश टेम्पलेट्स
	templates := map[string]string{
		"RAJASTHAN":   "राम राम सा! %s बेटा, कॉपी ते पैन चक्को। आज रो सवाल है: %s। लिखण के बाद 1 दबाओ।",
		"PUNJAB":      "सत श्री अकाल जी! %s पुत्तर, अपनी कापी ते पेन चक्को। आज दा सवाल है: %s। लिख के 1 दबाओ।",
		"TAMILNADU":   "வணக்கம் %s! உங்கள் நோட்டுப் புத்தகத்தை எடுங்கள். இன்றைய கேள்வி: %s. எழுதிய பின் 1 ஐ அழுத்தவும்.",
		"MAHARASHTRA": "नमस्कार %s बेटा! तुमची वही आणि पेन घ्या. आजचा प्रश्न आहे: %s. लिहून झाल्यावर 1 दाबा.",
		"BENGAL":      "নমস্কার %s! তোমার খাতা ও পেন নাও। আজকের প্রশ্ন হলো: %s। লেখার পর ১ প্রেস করো।",
	}

	// यदि राज्य का विशेष टेम्पलेट मिल जाए, तो वह उपयोग करें; अन्यथा डिफ़ॉल्ट हिंदी
	template, exists := templates[strings.ToUpper(stateCode)]
	if !exists {
		template = "नमस्ते! %s, कृपया अपनी कॉपी और पेन निकालें। आज का 15-मिनट का सवाल है: %s। लिखने के बाद 1 दबाएं।"
	}

	return fmt.Sprintf(template, studentName, mathProblem), nil
}

// GenerateDailyTaskSMS SMS द्वारा छात्र को दैनिक कार्य भेजने के लिए (भाषा के आधार पर हिंदी या अंग्रेजी)
func (f *FeaturePhoneEngine) GenerateDailyTaskSMS(studentName, problem, language string) string {
	if strings.ToUpper(language) == "ENGLISH" {
		return fmt.Sprintf("Anant Abhyaas: Hello %s, today's practice task: %s. Please complete and write it down within 15 minutes!", studentName, problem)
	}
	return fmt.Sprintf("अनंत अभ्यास: %s बेटा, आज का अभ्यास: %s। इसे 15 मिनट में लिखकर रखें!", studentName, problem)
}
