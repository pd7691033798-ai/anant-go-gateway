package featurephone

import (
	"database/sql"
	"fmt"
)

type FeaturePhoneEngine struct {
	db *sql.DB
}

func NewFeaturePhoneEngine(db *sql.DB) *FeaturePhoneEngine {
	return &FeaturePhoneEngine{db: db}
}

// बिना इंटरनेट वाले कीपैड फ़ोन के लिए IVR कॉल स्क्रिप्ट
func (f *FeaturePhoneEngine) GenerateIVRScript(studentName, dialectCode string, mathProblem string) string {
	if dialectCode == "BAGRI_PUNJABI_FUSION" {
		return fmt.Sprintf(
			"राम राम जी! %s बेटा, कॉपी ते पैन चक्को। आज रो सवाल है: %s। लिखण के बाद फ़ोन पर 1 दबाओ।",
			studentName, mathProblem,
		)
	}
	return fmt.Sprintf(
		"नमस्ते! %s, कृपया अपनी कॉपी निकालें। आज का 15-मिनट का सवाल है: %s। लिखने के बाद 1 दबाएं।",
		studentName, mathProblem,
	)
}

// SMS द्वारा दैनिक कार्य भेजना
func (f *FeaturePhoneEngine) GenerateDailyTaskSMS(studentName, problem string) string {
	return fmt.Sprintf("अनंत अभ्यास: %s बेटा, आज का अभ्यास: %s। इसे 15 मिनट में लिखकर रखें!", studentName, problem)
}
