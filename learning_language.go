package main

import (
	"errors"
	"fmt"
	"time"
)

type LanguageCategory string

const (
	CategoryRegional LanguageCategory = "REGIONAL" // 28 राज्य + अंग्रेजी (कुल 29)
	CategoryGlobal   LanguageCategory = "GLOBAL"   // 20 अंतरराष्ट्रीय + अंग्रेजी (कुल 21)
	CategoryFamily   LanguageCategory = "FAMILY"   // 2 बच्चों के लिए कॉम्बो पैक
)

type BillingCycle string

const (
	BillingMonthly BillingCycle = "MONTHLY"
	BillingAnnual  BillingCycle = "ANNUAL"
)

// अंग्रेजी के सभी 7 एक्सेंट
type EnglishAccent string

const (
	AccentAmerican   EnglishAccent = "AMERICAN"
	AccentBritish    EnglishAccent = "BRITISH"
	AccentAustralian EnglishAccent = "AUSTRALIAN"
	AccentCanadian   EnglishAccent = "CANADIAN"
	AccentIrish      EnglishAccent = "IRISH"
	AccentStandard   EnglishAccent = "STANDARD_ENGLISH"
	AccentIndian     EnglishAccent = "INDIAN_ENGLISH"
)

type LanguageConfig struct {
	PlanType          LanguageCategory
	Billing           BillingCycle
	SelectedLanguage  string
	SelectedAccent    EnglishAccent 
	DailyMinutes      int           
	ChildCount        int           // बच्चों की संख्या (1 या 2)
	LockInEndDate     time.Time     
	TotalFee          float64       
}

// नाम बदलकर LearningLanguage किया गया है
type LearningLanguage struct {
	RegionalLanguages []string
	GlobalLanguages   []string
	AvailableAccents  []EnglishAccent
}

func NewLearningLanguage() *LearningLanguage {
	// 28 भारतीय राज्य और 29वीं भाषा के रूप में English
	regional := []string{
		"Hindi", "Bengali", "Marathi", "Tamil", "Telugu", "Gujarati", "Kannada", 
		"Malayalam", "Punjabi", "Odia", "Assamese", "Kashmiri_Dogri", "Urdu", 
		"Sindhi", "Sanskrit", "Nepali", "Manipuri", "Konkani", "Bodo", "Santali", 
		"Maithili", "Bhojpuri_Rajasthani", "Mizo", "Khasi_Garo", "Kokborok", 
		"Nagamese", "Bhutia_Lepcha", "English",
	}

	// 20 अंतरराष्ट्रीय भाषाएं और 21वीं भाषा के रूप में English
	global := []string{
		"French", "Spanish", "German", "Mandarin", "Japanese", "Korean", 
		"Russian", "Italian", "Portuguese", "Arabic", "Persian", "Turkish", 
		"Dutch", "Swedish", "Greek", "Polish", "Vietnamese", "Thai", 
		"Indonesian", "Latin", "English",
	}

	accents := []EnglishAccent{
		AccentAmerican, AccentBritish, AccentAustralian, 
		AccentCanadian, AccentIrish, AccentStandard, AccentIndian,
	}

	return &LearningLanguage{
		RegionalLanguages: regional,
		GlobalLanguages:   global,
		AvailableAccents:  accents,
	}
}

func (ll *LearningLanguage) OnboardUser(plan LanguageCategory, billing BillingCycle, language string, accent EnglishAccent, minutes int, children int) (*LanguageConfig, error) {
	// दैनिक समय की जाँच (15 से 60 मिनट)
	if minutes < 15 || minutes > 60 {
		return nil, errors.New("दैनिक अभ्यास का समय 15 से 60 मिनट के बीच होना चाहिए")
	}

	var fee float64
	var childCount int

	switch plan {
	case CategoryRegional:
		childCount = 1
		if billing == BillingMonthly {
			fee = 399.0 // बेसिक प्लान मासिक शुल्क
		} else {
			fee = 3999.0 // बेसिक प्लान सालाना पैकेज
		}
		valid := false
		for _, lang := range ll.RegionalLanguages {
			if lang == language {
				valid = true
				break
			}
		}
		if !valid {
			return nil, errors.New("अमान्य क्षेत्रीय भाषा चयन")
		}

	case CategoryGlobal:
		childCount = 1
		if billing == BillingMonthly {
			fee = 599.0 // ग्लोबल प्लान मासिक शुल्क
		} else {
			fee = 5999.0 // ग्लोबल प्लान सालाना पैकेज
		}
		valid := false
		for _, lang := range ll.GlobalLanguages {
			if lang == language {
				valid = true
				break
			}
		}
		if !valid {
			return nil, errors.New("अमान्य अंतर्राष्ट्रीय भाषा चयन")
		}

	case CategoryFamily:
		childCount = 2 
		if children != 2 {
			return nil, errors.New("फैमिली पैक केवल 2 बच्चों के लिए है")
		}
		if billing == BillingMonthly {
			fee = 799.0 // फैमिली पैक मासिक शुल्क
		} else {
			fee = 7999.0 // फैमिली पैक सालाना पैकेज
		}
		valid := false
		for _, lang := range append(ll.RegionalLanguages, ll.GlobalLanguages...) {
			if lang == language {
				valid = true
				break
			}
		}
		if !valid {
			return nil, errors.New("अमान्य भाषा चयन")
		}

	default:
		return nil, errors.New("अमान्य प्लान प्रकार")
	}

	// यदि चुनी गई भाषा English है, तो 7 एक्सेंट में से एक चुनना अनिवार्य है
	if language == "English" {
		accentValid := false
		for _, acc := range ll.AvailableAccents {
			if acc == accent {
				accentValid = true
				break
			}
		}
		if !accentValid {
			return nil, errors.New("अंग्रेजी के लिए वैध एक्सेंट चुनना अनिवार्य है")
		}
	}

	// 7 दिन का लॉक-इन पीरियड सेट करना
	lockInEnd := time.Now().AddDate(0, 0, 7)

	return &LanguageConfig{
		PlanType:         plan,
		Billing:          billing,
		SelectedLanguage: language,
		SelectedAccent:   accent,
		DailyMinutes:     minutes,
		ChildCount:       childCount,
		LockInEndDate:    lockInEnd,
		TotalFee:         fee,
	}, nil
}

// दैनिक सीखने का चक्र (Daily Learning Cycle)
func RunDailySession(config *LanguageConfig) {
	vocabTime := config.DailyMinutes * 33 / 100 
	writeSpeakTime := config.DailyMinutes * 33 / 100 
	reviewTime := config.DailyMinutes - vocabTime - writeSpeakTime 

	fmt.Printf("--- दैनिक भाषा सत्र शुरू (%s) [बच्चे: %d] ---\n", config.SelectedLanguage, config.ChildCount)
	if config.SelectedLanguage == "English" {
		fmt.Printf("एक्सेंट मोड: %s\n", config.SelectedAccent)
	}
	fmt.Printf("1. शब्दकोश (Vocabulary): %d मिनट\n", vocabTime)
	fmt.Printf("2. लिखना और बोलना (Writing & Speaking): %d मिनट\n", writeSpeakTime)
	fmt.Printf("3. समीक्षा और फीडबैक (Review & Feedback): %d मिनट\n", reviewTime)
	fmt.Println("----------------------------------------")
}

func main() {
	learningLang := NewLearningLanguage()

	// उदाहरण परीक्षण: यूजर ने ग्लोबल प्लान चुना, सालाना बिलिंग चुनी, भाषा 'English' और 'INDIAN_ENGLISH' एक्सेंट चुना, समय 30 मिनट दिया
	userConfig, err := learningLang.OnboardUser(CategoryGlobal, BillingAnnual, "English", AccentIndian, 30, 1)
	if err != nil {
		fmt.Println("त्रुटि:", err)
		return
	}

	fmt.Printf("सफलतापूर्वक ऑनबोर्ड किया गया!\n प्लान प्रकार: %s\n बिलिंग चक्र: %s\n कुल शुल्क: ₹%.2f\n भाषा: %s\n एक्सेंट: %s\n बच्चों की संख्या: %d\n लॉक-इन समाप्त तिथि: %s\n\n", 
		userConfig.PlanType, userConfig.Billing, userConfig.TotalFee, userConfig.SelectedLanguage, userConfig.SelectedAccent, userConfig.ChildCount, userConfig.LockInEndDate.Format("2006-01-02"))

	RunDailySession(userConfig)
}

