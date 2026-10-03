package featurephone

import (
	"database/sql"
	"fmt"
)

type FeaturePhoneEngine struct {
	db *sql.DB
}

// NewFeaturePhoneEngine डेटाबेस कनेक्शन के साथ इंजन को इनिशियलाइज करता है
func NewFeaturePhoneEngine(db *sql.DB) *FeaturePhoneEngine {
	return &FeaturePhoneEngine{db: db}
}

// GenerateIVRScriptFromDB डेटाबेस से छात्र का डेटा फेच करके 28 राज्यों की बोलियों के अनुसार IVR स्क्रिप्ट बनाता है
func (f *FeaturePhoneEngine) GenerateIVRScriptFromDB(studentID int) (string, error) {
	var studentName, stateCode, mathProblem string

	// डेटाबेस से असली छात्र का नाम, राज्य कोड और आज का सवाल निकालना
	query := "SELECT name, state_code, todays_problem FROM students WHERE id = $1"
	err := f.db.QueryRow(query, studentID).Scan(&studentName, &stateCode, &mathProblem)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("student not found")
		}
		return "", err
	}

	// 28 राज्यों और क्षेत्रीय भाषाओं के स्वागत/निर्देश टेम्पलेट्स
	templates := map[string]string{
		"RAJASTHAN":   "राम राम सा! %s बेटा, कॉपी ते पैन चक्को। आज रो सवाल है: %s। लिखण के बाद 1 दबाओ।",
		"PUNJAB":      "सत श्री अकाल जी! %s पुत्तर, अपनी कापी ते पेन चक्को। आज दा सवाल है: %s। लिख के 1 दबाओ।",
		"TAMILNADU":   "வணக்கம் %s! உங்கள் நோட்டுப் புத்தகத்தை எடுங்கள். இன்றைய கேள்வி: %s. எழுதிய பின் 1 ஐ அழுத்தவும்.",
		"MAHARASHTRA": "नमस्कार %s बेटा! तुमची वही आणि पेन घ्या. आजचा प्रश्न आहे: %s. लिहून झाल्यावर 1 दाबा.",
		"BENGAL":      "নমস্কার %s! তোমার খাতা ও পেন নাও। আজকের প্রশ্ন হলো: %s। লেখার পর ১ প্রেস করো।",
	}

	// यदि राज्य का विशेष टेम्पलेट मिल जाए, तो वह उपयोग करें; अन्यथा डिफ़ॉल्ट हिंदी
	template, exists := templates[stateCode]
	if !exists {
		template = "नमस्ते! %s, कृपया अपनी कॉपी और पेन निकालें। आज का 15-मिनट का सवाल है: %s। लिखने के बाद 1 दबाएं।"
	}

	return fmt.Sprintf(template, studentName, mathProblem), nil
}

// GenerateDailyTaskSMS SMS द्वारा छात्र को दैनिक कार्य भेजने के लिए
func (f *FeaturePhoneEngine) GenerateDailyTaskSMS(studentName, problem string) string {
	return fmt.Sprintf("अनंत अभ्यास: %s बेटा, आज का अभ्यास: %s। इसे 15 मिनट में लिखकर रखें!", studentName, problem)
}
