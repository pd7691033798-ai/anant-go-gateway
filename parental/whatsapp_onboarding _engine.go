package parental

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type OnboardingStage string

const (
	// पन्ना 5: 9664006651 इनकमिंग व बायर फ़िल्टर
	StageGatewayFilter OnboardingStage = "STAGE_GATEWAY_FILTER"
	StageBuyerIntent   OnboardingStage = "STAGE_BUYER_INTENT"

	// पन्ना 1: Stage 1 (माता-पिता व सुरक्षा नियंत्रण)
	Stage1ParentChildName OnboardingStage = "STAGE_1_PARENT_CHILD_NAME"
	Stage1DutyDetails     OnboardingStage = "STAGE_1_DUTY_DETAILS"
	Stage1PhoneScreenTime OnboardingStage = "STAGE_1_PHONE_SCREEN_TIME"
	Stage1SecurityAuth    OnboardingStage = "STAGE_1_SECURITY_AUTH"
	Stage1OvertimeLock    OnboardingStage = "STAGE_1_OVERTIME_LOCK"
	Stage1DeviceSelection OnboardingStage = "STAGE_1_DEVICE_SELECTION"
	Stage1FocusMode       OnboardingStage = "STAGE_1_FOCUS_MODE"

	// पन्ना 2: Stage 2 (मल्टी-चाइल्ड, कक्षा, विषय, परीक्षा विंग + None, स्थान व साइलेंट ट्रैप)
	Stage2MultiChildSplit OnboardingStage = "STAGE_2_MULTI_CHILD_SPLIT"
	Stage2ClassGrade      OnboardingStage = "STAGE_2_CLASS_GRADE"
	Stage2SelectSubjects  OnboardingStage = "STAGE_2_SELECT_SUBJECTS"
	Stage2ExamWingTarget  OnboardingStage = "STAGE_2_EXAM_WING_TARGET"
	Stage2CurriculumTrap  OnboardingStage = "STAGE_2_CURRICULUM_TRAP"
	Stage2LocationAddress OnboardingStage = "STAGE_2_LOCATION_ADDRESS"
	Stage2StudentStatus   OnboardingStage = "STAGE_2_STUDENT_STATUS"

	// पन्ना 3: Stage 3 (लाइव स्पीड व विषय आधारित परीक्षा लूप)
	Stage3ParentTrustCheck OnboardingStage = "STAGE_3_PARENT_TRUST_CHECK"
	Stage3LiveExamRunning  OnboardingStage = "STAGE_3_LIVE_EXAM_RUNNING"

	// पन्ना 4: रिपोर्टिंग चक्र, मेरिट सर्टिफिकेट व 7-दिवसीय डेमो
	StageComplete OnboardingStage = "STAGE_COMPLETE"
)

type DiagnosticScoreLevel string

const (
	LevelPoor      DiagnosticScoreLevel = "Poor (स्तर 1 - 5/20 अंक)"
	LevelMedium    DiagnosticScoreLevel = "Medium (स्तर 2 - 8/20 अंक)"
	LevelGood      DiagnosticScoreLevel = "Good (स्तर 3 - 12/20 अंक)"
	LevelExcellent DiagnosticScoreLevel = "Excellent (स्तर 4 - 16/20 अंक)"
	LevelSuperHard DiagnosticScoreLevel = "Super Hard / S.E (स्तर 5 - 20/20 अंक)"
)

type QuestionItem struct {
	Subject  string
	Question string
	OptionA  string
	OptionB  string
	OptionC  string
	OptionD  string
	Correct  string // A, B, C, D
}

type CompleteParentProfile struct {
	ParentPhone         string
	ParentName          string
	ChildName           string
	DutyHours           string
	DutyShift           string
	DailyPhoneAllowance string
	AuthMode            string
	SecurityMasterPin   string
	AllowOvertime       bool
	ReportLockSeconds   int
	DeviceType          string
	FocusModeLevel      string
	MultiChildCount     int
	ClassGrade          int
	DetectedGrade       int
	SelectedSubjects    []string
	ExamGoalWing        string
	HasExamAddon        bool
	CurrentChapterName  string
	IsOverQualified     bool // कक्षा 5 vs 8 का साइलेंट ट्रैप फ्लैग
	BoardType           string
	StateLocation       string
	DistrictVillage     string
	BatchCode           string
	ReferralCode        string

	// लाइव टेस्ट स्टेट
	CurrentQuestionIdx int
	ActiveQuestions    []QuestionItem
	ScoreOutOf20       int
	TotalSolvingSec    int
	QuestionSentAt     time.Time
	IndicatorLevel     DiagnosticScoreLevel
	DemoID             string
	CertificateURL     string
	DemoStartDate      time.Time
	DemoEndDate        time.Time
	CurrentStage       OnboardingStage
	LastInteraction    time.Time
}

type FullWhatsAppEngine struct {
	profiles map[string]*CompleteParentProfile
	mu       sync.RWMutex
}

func NewFullWhatsAppEngine() *FullWhatsAppEngine {
	return &FullWhatsAppEngine{
		profiles: make(map[string]*CompleteParentProfile),
	}
}

// बच्चे के चुने विषयों, कक्षा और परीक्षा विंग के आधार पर वास्तविक प्रश्न बैंक
func generateExamQuestions(subjects []string, examWing string, grade int, isOverQualified bool) []QuestionItem {
	var bank []QuestionItem

	for _, sub := range subjects {
		s := strings.ToLower(sub)
		if strings.Contains(s, "गणित") || strings.Contains(s, "math") {
			if strings.Contains(examWing, "JEE") {
				bank = append(bank, QuestionItem{
					Subject:  "गणित (JEE विंग)",
					Question: "यदि f(x) = x² + 2x + 1 है, तो f'(2) का मान क्या होगा?",
					OptionA:  "4", OptionB: "5", OptionC: "6", OptionD: "8",
					Correct:  "C",
				})
			} else if grade > 5 || isOverQualified {
				bank = append(bank, QuestionItem{
					Subject:  "गणित (उच्च तार्किक)",
					Question: "एक रेलगाड़ी 60 किमी/घंटा की गति से चल रही है। 3 घंटे में वह कितनी दूरी तय करेगी?",
					OptionA:  "120 किमी", OptionB: "180 किमी", OptionC: "240 किमी", OptionD: "150 किमी",
					Correct:  "B",
				})
			} else {
				bank = append(bank, QuestionItem{
					Subject:  "गणित (प्राथमिक गणना)",
					Question: "एक डिब्बे में 45 टॉफियां हैं। 5 बच्चों में बराबर बांटने पर प्रत्येक को कितनी मिलेंगी?",
					OptionA:  "8", OptionB: "9", OptionC: "7", OptionD: "10",
					Correct:  "B",
				})
			}
		} else if strings.Contains(s, "रीजनिंग") || strings.Contains(s, "तर्क") || strings.Contains(s, "reasoning") {
			bank = append(bank, QuestionItem{
				Subject:  "मानसिक योग्यता (रीजनिंग)",
				Question: "श्रृंखला पूरी करें: 2, 4, 8, 16, ___?",
				OptionA:  "24", OptionB: "30", OptionC: "32", OptionD: "36",
				Correct:  "C",
			})
		} else if strings.Contains(s, "विज्ञान") || strings.Contains(s, "science") {
			bank = append(bank, QuestionItem{
				Subject:  "सामान्य विज्ञान",
				Question: "प्रकाश संश्लेषण (Photosynthesis) के लिए पौधों को किस गैस की आवश्यकता होती है?",
				OptionA:  "ऑक्सीजन", OptionB: "नाइट्रोजन", OptionC: "कार्बन डाइऑक्साइड", OptionD: "हाइड्रोजन",
				Correct:  "C",
			})
		} else if strings.Contains(s, "अंग्रेजी") || strings.Contains(s, "english") {
			bank = append(bank, QuestionItem{
				Subject:  "English Language",
				Question: "Choose the correct antonym of 'GENEROUS':",
				OptionA:  "Kind", OptionB: "Miserly", OptionC: "Polite", OptionD: "Brave",
				Correct:  "B",
			})
		}
	}

	if len(bank) == 0 {
		bank = append(bank, QuestionItem{
			Subject:  "सामान्य तर्कशक्ति",
			Question: "45 में 5 का भाग देने पर क्या उत्तर आएगा?",
			OptionA:  "8", OptionB: "9", OptionC: "7", OptionD: "10",
			Correct:  "B",
		})
	}
	return bank
}

func (e *FullWhatsAppEngine) ProcessMessage(phone, incomingText string) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	cleanPhone := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	text := strings.TrimSpace(incomingText)
	lower := strings.ToLower(text)
	upper := strings.ToUpper(text)
	now := time.Now()

	profile, exists := e.profiles[cleanPhone]
	if !exists {
		profile = &CompleteParentProfile{
			ParentPhone:       cleanPhone,
			CurrentStage:      StageGatewayFilter,
			ReportLockSeconds: 180,
			LastInteraction:   now,
		}
		e.profiles[cleanPhone] = profile
	}
	profile.LastInteraction = now

	switch profile.CurrentStage {

	// पन्ना 5: गेटवे 9664006651 बायर फ़िल्टर
	case StageGatewayFilter:
		if strings.Contains(upper, "REF_") || strings.Contains(upper, "PARTNER_") || strings.Contains(upper, "SHOP_") {
			parts := strings.Split(upper, " ")
			for _, part := range parts {
				if strings.HasPrefix(part, "REF_") || strings.HasPrefix(part, "PARTNER_") || strings.HasPrefix(part, "SHOP_") {
					profile.ReferralCode = strings.TrimSpace(part)
					break
				}
			}
		}

		if lower == "hi" || upper == "HI" || strings.Contains(lower, "start") || strings.Contains(lower, "नमस्ते") {
			profile.CurrentStage = StageBuyerIntent
			refBanner := ""
			if profile.ReferralCode != "" {
				refBanner = fmt.Sprintf("🎁 *रेफरल कोड %s मान्य हुआ!* (विशेष ₹100 छूट सक्रिय)\n\n", profile.ReferralCode)
			}
			return fmt.Sprintf("नमस्ते! *'अनंत अभ्यास'* (Royal FMC corporation) गेटवे 9664006651 पर आपका स्वागत है। 🎓\n\n%s"+
				"हमारा सिस्टम बच्चों के फोन से गेम/रील्स बंद करवाकर रोजाना *15 मिनट का अनुशासित नो-चीटिंग अभ्यास* करवाता है।\n\n"+
				"क्या आप अपने बच्चे के अध्ययन और परीक्षा सुधार के लिए गंभीर हैं?\n"+
				"👉 *[1] हाँ, बिल्कुल गंभीर हूँ*\n"+
				"👉 *[2] नहीं, सिर्फ देखने आया हूँ*", refBanner)
		}
		return "नमस्ते! 'अनंत अभ्यास' शुरू करने हेतु कृपया *Hi* लिखकर भेजें।"

	case StageBuyerIntent:
		if text == "1" || strings.Contains(lower, "हाँ") || strings.Contains(lower, "yes") || lower == "ha" {
			profile.CurrentStage = Stage1ParentChildName
			return "उत्कृष्ट! केवल समर्पित परिवारों के लिए ही यह अखाड़ा है। 🎯\n\n" +
				"📋 *Stage 1: फॉर्म 1 (माता-पिता व छात्र पहचान):*\n" +
				"कृपया अपना (माता/पिता) और बच्चे का नाम लिखकर भेजें:\n" +
				"_(उदा: राजेश शर्मा, आर्यन)_"
		}
		return "धन्यवाद! जब आप बच्चे की पढ़ाई के लिए तैयार हों, तब 'Hi' लिखकर पुनः शुरू कर सकते हैं।"

	// पन्ना 1: Stage 1 (अभिभावक नियंत्रण)
	case Stage1ParentChildName:
		parts := strings.Split(text, ",")
		if len(parts) >= 2 {
			profile.ParentName = strings.TrimSpace(parts[0])
			profile.ChildName = strings.TrimSpace(parts[1])
		} else {
			profile.ParentName = text
			profile.ChildName = "विद्यार्थी"
		}
		profile.CurrentStage = Stage1DutyDetails
		return fmt.Sprintf("धन्यवाद %s जी! ✅\n\n"+
			"💼 *Stage 1 - सवाल 1 व 2 (ड्यूटी व समय):*\n"+
			"1. आप रोजाना कितने घंटे ड्यूटी/काम पर रहते हैं?\n"+
			"2. आपकी शिफ्ट क्या है?\n"+
			"[1] दिन की ड्यूटी (Day Shift: 8-10 घंटे)\n"+
			"[2] रात की ड्यूटी (Night Shift)\n"+
			"[3] व्यापार / घर पर (Flexible)\n\n"+
			"👉 *1, 2 या 3 चुनकर भेजें:*", profile.ParentName)

	case Stage1DutyDetails:
		if text == "1" {
			profile.DutyShift = "Day"
		} else if text == "2" {
			profile.DutyShift = "Night"
		} else {
			profile.DutyShift = "Flexible"
		}
		profile.DutyHours = "8-10h"
		profile.CurrentStage = Stage1PhoneScreenTime
		return "⏳ *Stage 1 - सवाल 3 (फोन उपयोग सीमा):*\n" +
			"आप अपने बच्चे को पूरे दिन में कितने समय के लिए फोन देते हैं?\n" +
			"[1] 30 मिनट से 1 घंटा\n" +
			"[2] 1 से 2 घंटे\n" +
			"[3] 2 घंटे से अधिक (लत छुड़ाना अनिवार्य)\n\n" +
			"👉 *1, 2 या 3 लिखकर बताएं:*"

	case Stage1PhoneScreenTime:
		profile.DailyPhoneAllowance = text
		profile.CurrentStage = Stage1SecurityAuth
		return "🔐 *Stage 1 - सवाल 4 (सिक्योरिटी व मास्टर ऑथराइजेशन):*\n" +
			"बच्चे की सुरक्षा और ऐप लॉक के लिए क्या चुनना चाहते हैं?\n" +
			"[1] 6-Digital Master PIN\n" +
			"[2] Face / Fingerprint Scan\n\n" +
			"👉 *यदि 6-अंकीय पिन चाहते हैं, तो सीधे अपना 6 अंकों का मास्टर पिन लिखें (उदा: 982143):*"

	case Stage1SecurityAuth:
		cleanPin := strings.TrimSpace(text)
		if len(cleanPin) == 6 && isAllDigitsOnly(cleanPin) {
			profile.AuthMode = "6_DIGIT_PIN"
			profile.SecurityMasterPin = cleanPin
		} else {
			profile.AuthMode = "BIOMETRIC_OR_DEFAULT"
			profile.SecurityMasterPin = "889900"
		}
		profile.CurrentStage = Stage1OvertimeLock
		return "⏱️ *Stage 1 - सवाल 5 (अतिरिक्त समय व 3-मिनट रिपोर्ट लॉक):*\n" +
			"अभ्यास की सीमा समाप्त होने पर ऐप लॉक हो जाएगा। यदि बच्चा बाद में अतिरिक्त समय मांगे, तो क्या अनुमति देंगे?\n" +
			"[1] सख्त लॉक (कोई अतिरिक्त समय नहीं मिलेगा)\n" +
			"[2] 15 मिनट अतिरिक्त (केवल आपके मास्टर पिन से अनुमति)\n\n" +
			"💡 *नियम:* सीमा पूरी होते ही सिस्टम 3 मिनट के लिए लॉक होकर प्रोग्रेस रिपोर्ट दिखाएगा।\n" +
			"👉 *1 या 2 लिखकर चुनें:*"

	case Stage1OvertimeLock:
		profile.AllowOvertime = (text == "2")
		profile.CurrentStage = Stage1DeviceSelection
		return "📱 *Stage 1 - सवाल 6 (डिवाइस का प्रकार):*\n" +
			"बच्चा किस फोन पर अभ्यास करेगा?\n" +
			"[1] Android Smartphone (पूर्ण Kiosk स्क्रीन लॉक)\n" +
			"[2] iPhone (Apple iOS)\n\n" +
			"👉 *1 या 2 लिखकर बताएं:*"

	case Stage1DeviceSelection:
		if text == "1" {
			profile.DeviceType = "Android (Kiosk Mode)"
		} else {
			profile.DeviceType = "iPhone"
		}
		profile.CurrentStage = Stage1FocusMode
		return "🛡️ *Stage 1 (पैरेंटल फोकस मोड स्तर):*\n" +
			"अभ्यास के दौरान फोन पर कितना नियंत्रण चाहते हैं?\n" +
			"[1] Easy: सामान्य अभ्यास, स्क्रीन लॉक नहीं।\n" +
			"[2] Moderate: परिवार की कॉल्स चालू, बाकी ऐप बंद।\n" +
			"[3] Strict: स्क्रीन हार्डवेयर-लॉक (Kiosk), केवल आपातकालीन कॉल की अनुमति।\n\n" +
			"👉 *1, 2 या 3 चुनें:*"

	// पन्ना 2: Stage 2 (पारिवारिक स्लॉट, कक्षा, विषय, परीक्षा विंग + None, स्थान व साइलेंट ट्रैप)
	case Stage1FocusMode:
		profile.FocusModeLevel = text
		profile.CurrentStage = Stage2MultiChildSplit
		return "👨‍👩‍👧‍👦 *Stage 2 (मल्टी-चाइल्ड डिवाइस शेयरिंग - 1 फोन में 4 बच्चे):*\n" +
			"क्या आपके घर में 1 ही फोन पर कई बच्चे अभ्यास करेंगे?\n" +
			"[1] केवल 1 बच्चा है (15 मिनट सत्र)\n" +
			"[2] 2 बच्चे हैं (क्रमशः 15-15 मिनट स्लॉट विभाजन)\n" +
			"[3] 3 या 4 बच्चे हैं (पारिवारिक रोटेशन स्लॉट)\n\n" +
			"👉 *1, 2 या 3 लिखकर बताएं:*"

	case Stage2MultiChildSplit:
		c, _ := strconv.Atoi(text)
		if c <= 0 {
			c = 1
		}
		profile.MultiChildCount = c
		profile.CurrentStage = Stage2ClassGrade
		return "📚 *Stage 2 - सवाल (1) (विद्यार्थी की कक्षा):*\n" +
			"बच्चा वर्तमान में कौन सी कक्षा (Class 1 से 12) में पढ़ रहा है?\n" +
			"👉 *कक्षा संख्या लिखें (उदा: 5, 8, 10):*"

	case Stage2ClassGrade:
		g, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || g <= 0 {
			g = 6
		}
		profile.ClassGrade = g
		profile.DetectedGrade = g
		profile.CurrentStage = Stage2SelectSubjects
		return fmt.Sprintf("📖 *Stage 2 - सवाल (2) (कक्षा %d के चुने हुए विषय):*\n"+
			"बच्चा किन-किन विषयों का अभ्यास व टेस्ट देना चाहता है?\n"+
			"[1] गणित (Maths)\n"+
			"[2] विज्ञान (Science)\n"+
			"[3] रीजनिंग (Mental Ability)\n"+
			"[4] अंग्रेजी (English)\n\n"+
			"👉 *विषयों के नाम या नंबर कॉमा लगाकर लिखें (उदा: 1, 3 या गणित, रीजनिंग):*", profile.ClassGrade)

	case Stage2SelectSubjects:
		parts := strings.Split(text, ",")
		var subs []string
		for _, p := range parts {
			clean := strings.TrimSpace(p)
			if clean == "1" || strings.Contains(strings.ToLower(clean), "math") || strings.Contains(clean, "गणित") {
				subs = append(subs, "गणित")
			} else if clean == "2" || strings.Contains(strings.ToLower(clean), "science") || strings.Contains(clean, "विज्ञान") {
				subs = append(subs, "विज्ञान")
			} else if clean == "3" || strings.Contains(strings.ToLower(clean), "reasoning") || strings.Contains(clean, "रीजनिंग") {
				subs = append(subs, "रीजनिंग")
			} else if clean == "4" || strings.Contains(strings.ToLower(clean), "english") || strings.Contains(clean, "अंग्रेजी") {
				subs = append(subs, "अंग्रेजी")
			} else if clean != "" {
				subs = append(subs, clean)
			}
		}
		if len(subs) == 0 {
			subs = []string{"गणित", "रीजनिंग"}
		}
		profile.SelectedSubjects = subs
		profile.CurrentStage = Stage2ExamWingTarget

		return "🎯 *Stage 2 - सवाल (3) (लक्ष्य परीक्षा विंग का चयन):*\n\n" +
			"🏛️ *प्रतियोगी परीक्षा विंग्स:*\n" +
			"[A] नवोदय विद्यालय प्रवेश परीक्षा (JNVST)\n" +
			"[B] सैनिक स्कूल / मिलिट्री स्कूल (AISSEE)\n" +
			"[C] JEE Mains (NTA कंप्यूटर-बेस्ड टेस्ट पैटर्न)\n" +
			"[D] JEE Advanced (सुपर-हार्ड न्यूमेरिकल विंग)\n" +
			"[E] NDA / डिफ़ेंस फाउंडेशन विंग\n" +
			"[F] None (कोई विशेष परीक्षा नहीं, केवल स्कूल बोर्ड व सामान्य अभ्यास)\n\n" +
			"👉 *A, B, C, D, E या F लिखकर चुनें:*"

	case Stage2ExamWingTarget:
		choice := upper
		switch choice {
		case "A":
			profile.ExamGoalWing = "नवोदय विद्यालय (JNVST)"
		case "B":
			profile.ExamGoalWing = "सैनिक स्कूल (AISSEE)"
		case "C":
			profile.ExamGoalWing = "JEE Mains"
			profile.HasExamAddon = true
		case "D":
			profile.ExamGoalWing = "JEE Advanced"
			profile.HasExamAddon = true
		case "E":
			profile.ExamGoalWing = "NDA Foundation"
		default:
			profile.ExamGoalWing = "None (सामान्य स्कूल पाठ्यक्रम)"
			profile.HasExamAddon = false
		}
		profile.CurrentStage = Stage2CurriculumTrap
		return "📖 *Stage 2 - सवाल (4) (पाठ्यक्रम व चैप्टर पुष्टि):*\n" +
			"स्कूल में गणित (Maths) या विज्ञान में अभी कौन सा पाठ/चैप्टर चल रहा है?\n" +
			"_(उदा: भिन्न/Fractions, कोण, बीजगणित, रैखिक समीकरण लिख कर बताएं)_"

	// साइलेंट क्लास वेरिफिकेशन (कक्षा 5 vs 8 ट्रैप)
	case Stage2CurriculumTrap:
		profile.CurrentChapterName = text
		lowerChap := strings.ToLower(text)
		if profile.ClassGrade <= 5 {
			if strings.Contains(lowerChap, "algebra") || strings.Contains(lowerChap, "बीजगणित") ||
				strings.Contains(lowerChap, "linear") || strings.Contains(lowerChap, "वर्गमूल") ||
				strings.Contains(lowerChap, "triangles") {
				profile.IsOverQualified = true
				profile.DetectedGrade = 8 // वास्तविक स्तर 8वीं पकड़ा गया
			}
		}
		profile.CurrentStage = Stage2LocationAddress
		return "📍 *Stage 2 - सवाल (5) (स्थान व पता):*\n" +
			"आप किस राज्य, जिले, तहसील और गांव से हैं?\n" +
			"_(उदा: राजस्थान, जयपुर, बस्सी)_"

	case Stage2LocationAddress:
		parts := strings.Split(text, ",")
		profile.StateLocation = text
		if len(parts) >= 2 {
			profile.DistrictVillage = strings.TrimSpace(parts[1])
		}
		profile.CurrentStage = Stage2StudentStatus
		return "🔄 *Stage 2 - सवाल (6) (सिस्टम बैच पहचान):*\n" +
			"क्या बच्चा पहली बार सिस्टम पर टेस्ट दे रहा है?\n" +
			"[1] नया बच्चा है (New Enrolled)\n" +
			"[2] पुराना छात्र है (प्रमोटेड बैच)\n\n" +
			"👉 *1 या 2 लिखकर बताएं:*"

	// पन्ना 3: Stage 3 (माता-पिता की परीक्षा + चुने विषयों का लाइव टेस्ट लूप)
	case Stage2StudentStatus:
		if text == "2" {
			profile.BatchCode = "BATCH-PROMOTED-2026"
		} else {
			profile.BatchCode = "BATCH-NEW-2026"
		}
		profile.CurrentStage = Stage3ParentTrustCheck
		return "🧪 *Stage 3 (सिस्टम विशेषता व लाइव डायग्नोस्टिक टेस्ट):*\n" +
			"अनंत अभ्यास के 3 कड़े नियम:\n" +
			"• 🚫 3-सेकंड नो-तुक्का इंजन (तुक्केबाजी पर तुरंत टेस्ट लॉक)\n" +
			"• 🔒 हार्डवेयर Kiosk (स्क्रीनशॉट व अन्य ऐप्स पूरी तरह ब्लॉक)\n" +
			"• ⏱️ लाइव स्पीड व रिस्पॉन्स टाइमर\n\n" +
			"💡 *सिस्टम डेमो देने से पहले अब बच्चे के चुने हुए विषयों का लाइव टेस्ट शुरू होगा!*\n" +
			"क्या बच्चा टेस्ट के लिए तैयार है? लिखें: *हाँ / नहीं*"

	case Stage3ParentTrustCheck:
		if strings.Contains(lower, "हाँ") || strings.Contains(lower, "yes") || lower == "ha" {
			profile.ActiveQuestions = generateExamQuestions(profile.SelectedSubjects, profile.ExamGoalWing, profile.ClassGrade, profile.IsOverQualified)
			profile.CurrentQuestionIdx = 0
			profile.ScoreOutOf20 = 0
			profile.TotalSolvingSec = 0
			profile.CurrentStage = Stage3LiveExamRunning
			profile.QuestionSentAt = time.Now()

			firstQ := profile.ActiveQuestions[0]
			return fmt.Sprintf("📝 *लाइव टेस्ट शुरू (प्रश्न 1/%d)*\n"+
				"• छात्र: %s | कक्षा: %d | परीक्षा: %s\n"+
				"• विषय: %s\n"+
				"━━━━━━━━━━━━━━━━━━━━\n"+
				"❓ %s\n\n"+
				"[A] %s\n[B] %s\n[C] %s\n[D] %s\n"+
				"━━━━━━━━━━━━━━━━━━━━\n"+
				"⏱️ *टाइमर चालू है।* सही विकल्प (A, B, C या D) तुरंत भेजें:",
				len(profile.ActiveQuestions), profile.ChildName, profile.ClassGrade, profile.ExamGoalWing,
				firstQ.Subject, firstQ.Question, firstQ.OptionA, firstQ.OptionB, firstQ.OptionC, firstQ.OptionD)
		}
		return "तैयार होने पर 'हाँ' लिखकर भेजें।"

	// लाइव टेस्ट क्वेश्चन लूप
	case Stage3LiveExamRunning:
		timeTaken := int(time.Since(profile.QuestionSentAt).Seconds())
		profile.TotalSolvingSec += timeTaken

		// स्पीड ट्रैप: यदि 5वीं का दावा था लेकिन 2-स्टेप सवाल 6 सेकंड से कम में दे दिया
		if profile.ClassGrade <= 5 && timeTaken < 7 && !profile.IsOverQualified {
			profile.IsOverQualified = true
			profile.DetectedGrade = 8
		}

		currQ := profile.ActiveQuestions[profile.CurrentQuestionIdx]
		if upper == currQ.Correct {
			profile.ScoreOutOf20 += (20 / len(profile.ActiveQuestions))
		}

		profile.CurrentQuestionIdx++

		// अगर और सवाल बचे हैं तो अगला सवाल
		if profile.CurrentQuestionIdx < len(profile.ActiveQuestions) {
			profile.QuestionSentAt = time.Now()
			nextQ := profile.ActiveQuestions[profile.CurrentQuestionIdx]
			return fmt.Sprintf("✅ उत्तर दर्ज हुआ! (%d सेकंड)\n\n"+
				"📝 *प्रश्न %d/%d (%s)*\n"+
				"━━━━━━━━━━━━━━━━━━━━\n"+
				"❓ %s\n\n"+
				"[A] %s\n[B] %s\n[C] %s\n[D] %s\n"+
				"━━━━━━━━━━━━━━━━━━━━\n"+
				"👉 सही विकल्प (A, B, C या D) भेजें:",
				timeTaken, profile.CurrentQuestionIdx+1, len(profile.ActiveQuestions), nextQ.Subject,
				nextQ.Question, nextQ.OptionA, nextQ.OptionB, nextQ.OptionC, nextQ.OptionD)
		}

		// पन्ना 4: मूल्यांकन, इंडिकेटर इंडेक्स, मेरिट सर्टिफिकेट व 7-दिवसीय डेमो
		profile.CurrentStage = StageComplete
		score := profile.ScoreOutOf20
		if score <= 5 {
			profile.IndicatorLevel = LevelPoor
		} else if score <= 8 {
			profile.IndicatorLevel = LevelMedium
		} else if score <= 12 {
			profile.IndicatorLevel = LevelGood
		} else if score <= 16 {
			profile.IndicatorLevel = LevelExcellent
		} else {
			profile.IndicatorLevel = LevelSuperHard
		}

		last4 := cleanPhone
		if len(cleanPhone) >= 4 {
			last4 = cleanPhone[len(cleanPhone)-4:]
		}
		profile.DemoID = fmt.Sprintf("DEMO-2026-%s", last4)
		profile.DemoStartDate = now
		profile.DemoEndDate = now.AddDate(0, 0, 7) // 7-दिन का फ्री डेमो

		// डिजिटल सर्टिफिकेट यूआरएल (नीचे दिए HTML हैंडलर से सर्व होगा)
		profile.CertificateURL = fmt.Sprintf("https://anantabhyas.com/cert/%s", profile.DemoID)

		auditReport := "सत्यापित (कक्षा और रिस्पॉन्स समय संतुलित)"
		if profile.IsOverQualified {
			auditReport = fmt.Sprintf("⚡ असाधारण गति (%ds)! बच्चे की क्षमता कक्षा %d से आगे (कक्षा %d स्तर) पाई गई। सिस्टम ने टेस्ट स्तर स्वतः अपग्रेड किया।",
				profile.TotalSolvingSec, profile.ClassGrade, profile.DetectedGrade)
		}

		totalAmount := 399.0
		if profile.HasExamAddon {
			totalAmount = 1098.0 // ₹399 + ₹799 - ₹100 छूट
		}

		vpa := "9664006651@ptsbi"
		merchant := "Royal FMC corporation"
		upiLink := fmt.Sprintf("upi://pay?pa=%s&pn=%s&am=%.2f&cu=INR&tn=%s", vpa, url.QueryEscape(merchant), totalAmount, url.QueryEscape(profile.DemoID))
		qrCodeLink := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=%s", url.QueryEscape(upiLink))

		return fmt.Sprintf("🏁 *लाइव टेस्ट पूर्ण! छात्र मूल्यांकन व प्रोग्रेस रिपोर्ट*\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"• छात्र: *%s* | अभिभावक: *%s*\n"+
			"• कक्षा: *Class %d* | विषय: *%s*\n"+
			"• परीक्षा विंग: *%s*\n"+
			"• स्पीड ऑडिट: *%s*\n"+
			"• टेस्ट स्कोर: *%d / 20 अंक*\n"+
			"• इंडिकेटर स्तर: *%s*\n"+
			"• 🆔 डेमो ID: *%s*\n"+
			"━━━━━━━━━━━━━━━━━━━━\n\n"+
			"🎖️ *आधिकारिक डिजिटल मेरिट सर्टिफिकेट (Certificate of Merit):*\n"+
			"👉 %s\n"+
			"_(माता-पिता इसे डाउनलोड व WhatsApp स्टेटस पर शेयर कर सकते हैं)_\n\n"+
			"📈 *पैरेंट रिपोर्टिंग चक्र (पन्ना 4):*\n"+
			"• *7-दिन की कंपेयर रिपोर्ट:* पहले व 7वें दिन की प्रगति का तुलनात्मक डेटा।\n"+
			"• *21-दिन की मासिक रिपोर्ट:* 80%%+ उपस्थिति पर विस्तृत परफ़ॉर्मेंस शीट।\n\n"+
			"🎉 *सत्यापन सफल! 7-दिन का फ्री सिस्टम डेमो सक्रिय हो गया है!*\n\n"+
			"📲 *यूज़र ऐप डाउनलोड लिंक:*\n"+
			"👉 https://anantabhyas.com/download/user-app\n\n"+
			"🔐 *लॉगिन निर्देश:*\n"+
			"1. मोबाइल नंबर (+91 %s) दर्ज करें।\n"+
			"2. मास्टर पिन: *%s* दर्ज करें।\n"+
			"3. 15 मिनट पूरे होते ही सिस्टम 3 मिनट के लिए लॉक होकर रिपोर्ट दिखाएगा।\n\n"+
			"💳 *7-दिन बाद UPI ऑटो-पे सेटअप QR:* %s",
			profile.ChildName, profile.ParentName, profile.ClassGrade, strings.Join(profile.SelectedSubjects, ", "),
			profile.ExamGoalWing, auditReport, profile.ScoreOutOf20, profile.IndicatorLevel,
			profile.DemoID, profile.CertificateURL, cleanPhone, profile.SecurityMasterPin, qrCodeLink)

	case StageComplete:
		return fmt.Sprintf("नमस्ते %s जी! %s का 7-दिवसीय डेमो सक्रिय है (ID: %s)। ऐप खोलकर आज का 15-मिनट अभ्यास पूरा करें। सर्टिफिकेट लिंक: %s",
			profile.ParentName, profile.ChildName, profile.DemoID, profile.CertificateURL)
	}

	return "नमस्ते! अभ्यास शुरू करने हेतु *Hi* भेजें।"
}

// एडमिन या सर्टिफिकेट व्यूअर के लिए प्रोफाइल प्राप्त करना
func (e *FullWhatsAppEngine) GetProfileByDemoID(demoID string) *CompleteParentProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, p := range e.profiles {
		if p.DemoID == demoID {
			return p
		}
	}
	return nil
}

func isAllDigitsOnly(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
