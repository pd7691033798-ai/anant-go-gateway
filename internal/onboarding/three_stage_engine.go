package onboarding

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type FormStage string

const (
	Stage1_ChildName       FormStage = "S1_CHILD_NAME"
	Stage1_ParentName      FormStage = "S1_PARENT_NAME"
	Stage1_LanguageSelect  FormStage = "S1_LANGUAGE_SELECT"
	Stage2_DutyHours       FormStage = "S2_DUTY_HOURS"
	Stage2_DutyShift       FormStage = "S2_DUTY_SHIFT"
	Stage2_PhoneTimeDaily  FormStage = "S2_PHONE_TIME_DAILY"
	Stage2_Biometrics      FormStage = "S2_BIOMETRICS_OPT"
	Stage2_ParentPIN       FormStage = "S2_PARENT_PIN"
	Stage2_DeviceType      FormStage = "S2_DEVICE_TYPE"
	Stage2_ExtraTime       FormStage = "S2_EXTRA_TIME_ALLOWANCE"
	Stage2_ReportConsent   FormStage = "S2_REPORT_CONSENT"
	Stage2_Suggestion      FormStage = "S2_SUGGESTION_BOX"
	Stage3_ClassLevel      FormStage = "S3_CLASS_LEVEL"
	Stage3_ClassVerifyMode FormStage = "S3_CLASS_VERIFY_MODE"
	Stage3_SchoolName      FormStage = "S3_SCHOOL_NAME"
	Stage3_TotalSubjects   FormStage = "S3_TOTAL_SUBJECTS"
	Stage3_BoardName       FormStage = "S3_BOARD_NAME"
	Stage3_PincodeGeo      FormStage = "S3_PINCODE_GEO"
	Stage3_TotalSiblings   FormStage = "S3_TOTAL_SIBLINGS"
	Stage_Completed        FormStage = "STAGE_COMPLETED"
)

type ThreeStageFormEngine struct {
	db *sql.DB
}

func NewThreeStageFormEngine(db *sql.DB) *ThreeStageFormEngine {
	return &ThreeStageFormEngine{db: db}
}

func (t *ThreeStageFormEngine) ProcessUserInput(parentPhone, input string) string {
	parentPhone = strings.TrimPrefix(parentPhone, "+91")
	text := strings.TrimSpace(input)

	var currentStage FormStage
	query := `SELECT COALESCE(current_stage, 'S1_CHILD_NAME') FROM onboarding_sessions WHERE phone = $1`
	err := t.db.QueryRow(query, parentPhone).Scan(&currentStage)

	if err != nil {
		t.initSession(parentPhone)
		return "🙏 *नमस्कार! अनंत अभ्यास में आपका स्वागत है।*\n\n👉 सबसे पहले *बच्चे (छात्र) का शुभ नाम* बताएं:"
	}

	switch currentStage {
	// ==================== स्टेज 1: परिचय व 28-राज्यों की भाषा चयन ====================
	case Stage1_ChildName:
		t.updateField(parentPhone, "child_name", text, Stage1_ParentName)
		return fmt.Sprintf("बेटा %s के *पिता/अभिभावक का पूरा नाम* बताएं:", text)

	case Stage1_ParentName:
		t.updateField(parentPhone, "parent_name", text, Stage1_LanguageSelect)
		return strings.Join([]string{
			"🌐 *अपने राज्य की भाषा या स्थानीय बोली चुनें (Pan-India 28 States):*",
			"1. हिन्दी (Standard Hindi)",
			"2. राजस्थानी / मारवाड़ी / बागड़ी (Rajasthan)",
			"3. हरियाणवी / जाटू (Haryana)",
			"4. भोजपुरी (Bihar / UP)",
			"5. मैथिली / अंगिका (Bihar)",
			"6. पंजाबी (Punjab)",
			"7. छत्तीसगढ़ी (Chhattisgarh)",
			"8. पहाड़ी / गढ़वाली / कुमाऊँनी (HP / UK)",
			"9. বাংলা / Bengali (West Bengal)",
			"10. मराठी (Maharashtra)",
			"11. తెలుగు (Telugu - AP / Telangana)",
			"12. தமிழ் (Tamil - Tamil Nadu)",
			"13. ગુજરાતી (Gujarati - Gujarat)",
			"14. ಕನ್ನಡ (Kannada - Karnataka)",
			"15. മലയാളം (Malayalam - Kerala)",
			"16. ओड़िया (Odisha)",
			"17. असमिया (Assam / North-East)",
			"18. अन्य राज्य बोलियां / English",
			"\n*(कृपया विकल्प नंबर 1 से 18 तक भेजें)*",
		}, "\n")

	case Stage1_LanguageSelect:
		dialect := "HINDI"
		switch text {
		case "2":
			dialect = "RAJASTHANI_BAGRI"
		case "3":
			dialect = "HARYANVI"
		case "4":
			dialect = "BHOJPURI"
		case "5":
			dialect = "MAITHILI"
		case "6":
			dialect = "PUNJABI"
		case "7":
			dialect = "CHHATTISGARHI"
		case "8":
			dialect = "PAHADI_GARHWALI"
		case "9":
			dialect = "BENGALI"
		case "10":
			dialect = "MARATHI"
		case "11":
			dialect = "TELUGU"
		case "12":
			dialect = "TAMIL"
		case "13":
			dialect = "GUJARATI"
		case "14":
			dialect = "KANNADA"
		case "15":
			dialect = "MALAYALAM"
		case "16":
			dialect = "ODIA"
		case "17":
			dialect = "ASSAMESE"
		case "18":
			dialect = "ENGLISH"
		default:
			dialect = "HINDI"
		}
		t.updateField(parentPhone, "preferred_dialect", dialect, Stage2_DutyHours)
		return "आप रोज़ाना कितने घंटे काम/ड्यूटी पर रहते हैं? (उदा. 8, 10 या 12 घंटे):"

	// ==================== स्टेज 2: पेरेंट कंट्रोल्स ====================
	case Stage2_DutyHours:
		t.updateField(parentPhone, "duty_hours", text, Stage2_DutyShift)
		return "आपकी ड्यूटी किस समय रहती है?\n1. दिन में\n2. रात में"

	case Stage2_DutyShift:
		t.updateField(parentPhone, "duty_shift", text, Stage2_PhoneTimeDaily)
		return "आप बच्चे को पढ़ाई के लिए दिन में कुल कितने समय फोन दे सकते हैं? (उदा. 60 मिनट):"

	case Stage2_PhoneTimeDaily:
		t.updateField(parentPhone, "phone_time_minutes", text, Stage2_Biometrics)
		return "🔒 *सुरक्षा:* क्या आप ऐप लॉक के लिए Face/Fingerprint स्कैन लगाना चाहेंगे?\n1. हाँ\n2. नहीं"

	case Stage2_Biometrics:
		if text == "2" || strings.ToLower(text) == "नहीं" {
			t.updateField(parentPhone, "biometrics_opted", "FALSE", Stage2_ParentPIN)
			return "कृपया अपना *6-अंकीय पेरेंट सुरक्षा PIN* सेट करें (उदा. 123456):"
		}
		t.updateField(parentPhone, "biometrics_opted", "TRUE", Stage2_DeviceType)
		return "घर में अभ्यास के लिए कौन सा फोन उपयोग करेंगे?\n1. Android\n2. iPhone\n3. Feature Phone (कीपैड)"

	case Stage2_ParentPIN:
		t.updateField(parentPhone, "parent_pin", text, Stage2_DeviceType)
		return "PIN सेट हुआ। अभ्यास के लिए कौन सा फोन उपयोग करेंगे?\n1. Android\n2. iPhone\n3. Feature Phone (कीपैड)"

	case Stage2_DeviceType:
		dev := "ANDROID"
		if text == "2" {
			dev = "IPHONE"
		} else if text == "3" {
			dev = "FEATURE_PHONE"
		}
		t.updateField(parentPhone, "device_type", dev, Stage2_ExtraTime)
		return "समय सीमा पूरी होने के बाद क्या आप अतिरिक्त समय देंगे?\n1. बिल्कुल नहीं (0 मिनट)\n2. 5 मिनट\n3. 10 मिनट"

	case Stage2_ExtraTime:
		t.updateField(parentPhone, "extra_time_allowed", text, Stage2_ReportConsent)
		return "क्या अभ्यास के बाद दैनिक प्रोग्रेस रिपोर्ट WhatsApp पर देखना चाहेंगे?\n1. हाँ\n2. नहीं"

	case Stage2_ReportConsent:
		t.updateField(parentPhone, "report_consent", text, Stage2_Suggestion)
		return "📝 *सुझाव बॉक्स:* हमारे लिए कोई सुझाव हो तो लिखकर भेजें (या 'कोई नहीं' लिखें):"

	case Stage2_Suggestion:
		t.updateField(parentPhone, "suggestion_text", text, Stage3_ClassLevel)
		return "बच्चा कौन सी कक्षा में पढ़ता है? (कक्षा *1 से 12* में से केवल नंबर दर्ज करें, उदा. 5, 8, 10):"

	// ==================== स्टेज 3: कक्षा चयन व सत्यापन ====================
	case Stage3_ClassLevel:
		classNum, err := strconv.Atoi(text)
		if err != nil || classNum < 1 || classNum > 12 {
			return "⚠️ केवल *1 से 12* के बीच का नंबर लिखें (उदा. 5, 10 या 12):"
		}
		t.updateField(parentPhone, "class_level", text, Stage3_ClassVerifyMode)
		nextClass := classNum + 1
		if classNum == 12 {
			nextClass = 12
		}
		return fmt.Sprintf("बेटा, आपने कक्षा %d चुनी है। कृपया अपनी वास्तविक स्थिति चुनें:\n\n1. मैं अभी स्कूल में कक्षा %d में ही पढ़ रहा हूँ\n2. मैं कक्षा %d पास कर चुका हूँ, अब कक्षा %d की तैयारी करनी है\n3. मैं कक्षा %d में हूँ, लेकिन मुझे अगली कक्षा (%d) की अग्रिम जानकारी/गाइडेंस भी चाहिए (₹0 फ्री)\n\n*(विकल्प 1, 2 या 3 लिखकर भेजें)*",
			classNum, classNum, classNum-1, classNum, classNum, nextClass)

	case Stage3_ClassVerifyMode:
		learningMode := "CURRENT_GRADE_ONLY"
		if text == "2" {
			learningMode = "PREPARING_NEW_CLASS"
		} else if text == "3" {
			learningMode = "ADVANCE_CURIOUS_USER"
		}
		t.updateField(parentPhone, "learning_intent", learningMode, Stage3_SchoolName)
		return "बच्चे के स्कूल/विद्यालय का नाम क्या है?"

	case Stage3_SchoolName:
		t.updateField(parentPhone, "school_name", text, Stage3_TotalSubjects)
		return "बच्चे के मुख्य विषय कितने हैं? (उदा. 3, 5, 6):"

	case Stage3_TotalSubjects:
		t.updateField(parentPhone, "total_subjects", text, Stage3_BoardName)
		return "बच्चे का बोर्ड कौन सा है?\n1. RBSE (राजस्थान बोर्ड)\n2. CBSE\n3. अन्य राज्य बोर्ड"

	case Stage3_BoardName:
		t.updateField(parentPhone, "board_name", text, Stage3_PincodeGeo)
		return "📍 अपने इलाके का *6-अंकीय पिनकोड* दर्ज करें (उदा. 335001):"

	case Stage3_PincodeGeo:
		t.updateField(parentPhone, "pincode", text, Stage3_TotalSiblings)
		return "घर में इस फोन पर कुल कितने बच्चे अभ्यास करेंगे? (उदा. 1, 2, 3, 4):"

	case Stage3_TotalSiblings:
		siblingsCount := 1
		fmt.Sscanf(text, "%d", &siblingsCount)
		badgeCode := fmt.Sprintf("ABHYAS-INDIA-2026-%s", parentPhone[len(parentPhone)-4:])
		t.finalizeProfile(parentPhone, badgeCode, siblingsCount)

		timePerChild := 60 / siblingsCount
		if siblingsCount > 1 {
			return fmt.Sprintf("🎉 *पंजीकरण पूर्ण! स्थायी लाइफटाइम बैच कोड:* `%s`\n\n📱 *मल्टी-चाइल्ड टाइम-शेयरिंग सक्रिय:*\nसभी %d बच्चों को बराबर *%d-%d मिनट* का ऑटो स्लॉट मिला है।\n\n7-दिन का फ्री डेमो सक्रिय है। अभ्यास के लिए 9664006651 पर मिसकॉल दें।",
				badgeCode, siblingsCount, timePerChild, timePerChild)
		}
		return fmt.Sprintf("🎉 *पंजीकरण पूर्ण!*\n\n🆔 *लाइफटाइम बैच कोड (UID):* `%s`\n🎁 *7-दिन का फ्री डेमो सक्रिय है (कक्षा 1 से 12 तक)।*\n\nअभ्यास के लिए 9664006651 पर मिसकॉल दें।", badgeCode)
	}

	return "कृपया सही विकल्प चुनकर उत्तर दें।"
}

func (t *ThreeStageFormEngine) initSession(phone string) {
	_, _ = t.db.Exec(`INSERT INTO onboarding_sessions (phone, current_stage, created_at) VALUES ($1, 'S1_CHILD_NAME', NOW()) ON CONFLICT (phone) DO UPDATE SET current_stage = 'S1_CHILD_NAME'`, phone)
}

func (t *ThreeStageFormEngine) updateField(phone, column, value string, nextStage FormStage) {
	query := fmt.Sprintf(`UPDATE onboarding_sessions SET %s = $1, current_stage = $2 WHERE phone = $3`, column)
	_, _ = t.db.Exec(query, value, nextStage, phone)
}

func (t *ThreeStageFormEngine) finalizeProfile(phone, badgeCode string, siblings int) {
	query := `
		UPDATE onboarding_sessions SET current_stage = 'STAGE_COMPLETED', total_children = $2 WHERE phone = $1;
		INSERT INTO users (phone, uid_badge, is_active, demo_days_used, total_children)
		VALUES ($1, $3, TRUE, 0, $2)
		ON CONFLICT (phone) DO UPDATE SET uid_badge = $3, total_children = $2;`
	_, _ = t.db.Exec(query, phone, siblings, badgeCode)
}
