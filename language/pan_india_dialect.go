package language

import (
	"strings"
)

type RegionalProfile struct {
	State       string `json:"state"`
	District    string `json:"district"`
	DialectName string `json:"dialect_name"`
	ToneHint    string `json:"tone_hint"`
}

type PanIndiaDialectService struct {
	// कस्टम यूजर-डिफ़ाइंड बोलियों का इन-मेमोरी / रनटाइम मैप
	customDialectMap map[string]RegionalProfile
}

func NewPanIndiaDialectService() *PanIndiaDialectService {
	return &PanIndiaDialectService{
		customDialectMap: make(map[string]RegionalProfile),
	}
}

// cleanString स्पेस हटाकर लोअरकेस बनाता है
func cleanString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// RegisterCustomUserDialect यदि यूजर खुद अपनी अनूठी भाषा लिखता है तो उसे रजिस्टर करता है
func (p *PanIndiaDialectService) RegisterCustomUserDialect(userLanguageInput, userState string) RegionalProfile {
	cleanLang := cleanString(userLanguageInput)
	cleanSt := cleanString(userState)

	profile := RegionalProfile{
		State:       userState,
		District:    "Custom/Self-Declared",
		DialectName: strings.ToUpper(cleanLang),
		ToneHint:    "बहुत बढ़िया / शाबाश बेटा",
	}

	// कुछ आम कस्टम इनपुट्स के लिए स्मार्ट टोन मैपिंग
	switch {
	case strings.Contains(cleanLang, "kashmiri") || strings.Contains(cleanLang, "कश्मीरी"):
		profile.DialectName = "KASHMIRI"
		profile.ToneHint = "वारियाह असल / शाबाश"
	case strings.Contains(cleanLang, "dogri") || strings.Contains(cleanLang, "डोगरी"):
		profile.DialectName = "DOGRI"
		profile.ToneHint = "बड़ा शैल्ल काम / शाबाश"
	case strings.Contains(cleanLang, "ladakhi") || strings.Contains(cleanLang, "bhoti") || strings.Contains(cleanLang, "लद्दाखी"):
		profile.DialectName = "LADAKHI"
		profile.ToneHint = "Julley! खूब बढ़िया काम"
	case strings.Contains(cleanLang, "mizo") || strings.Contains(cleanLang, "मिज़ो"):
		profile.DialectName = "MIZO"
		profile.ToneHint = "Tha lutuk / Very Good"
	case strings.Contains(cleanLang, "manipuri") || strings.Contains(cleanLang, "meitei"):
		profile.DialectName = "MANIPURI"
		profile.ToneHint = "Yamna phajei / शाबाश"
	case strings.Contains(cleanLang, "nepali") || strings.Contains(cleanLang, "नेपाली"):
		profile.DialectName = "NEPALI"
		profile.ToneHint = "धेरै राम्रो / शाबाश"
	case strings.Contains(cleanLang, "santhali") || strings.Contains(cleanLang, "संथाली"):
		profile.DialectName = "SANTHALI"
		profile.ToneHint = "Adi bes / शाबाश"
	}

	p.customDialectMap[cleanLang+"_"+cleanSt] = profile
	return profile
}

// ResolveRegionalDialect राज्य, ज़िले और स्वतः-पहचाने गए इनपुट से बोली तय करता है
func (p *PanIndiaDialectService) ResolveRegionalDialect(state, district string) RegionalProfile {
	s := cleanString(state)
	d := cleanString(district)

	// ==================== 1. जम्मू, कश्मीर और लद्दाख ====================
	if strings.Contains(s, "kashmir") || strings.Contains(s, "jammu") {
		if strings.Contains(d, "jammu") || strings.Contains(d, "samba") || strings.Contains(d, "kathua") || strings.Contains(d, "udhampur") {
			return RegionalProfile{State: state, District: district, DialectName: "DOGRI", ToneHint: "बड़ा शैल्ल काम / शाबाश"}
		}
		return RegionalProfile{State: state, District: district, DialectName: "KASHMIRI", ToneHint: "वारियाह असल / शाबाश"}
	}
	if strings.Contains(s, "ladakh") || strings.Contains(d, "leh") || strings.Contains(d, "kargil") {
		return RegionalProfile{State: state, District: district, DialectName: "LADAKHI", ToneHint: "Julley! खूब बढ़िया काम / शाबाश"}
	}

	// ==================== 2. उत्तर-पूर्व (North-East 8 States) ====================
	if strings.Contains(s, "assam") {
		return RegionalProfile{State: state, District: district, DialectName: "ASSAMESE", ToneHint: "বৰ ধুনীয়া হৈছে / শাবাশ"}
	}
	if strings.Contains(s, "manipur") {
		return RegionalProfile{State: state, District: district, DialectName: "MANIPURI", ToneHint: "Yamna phajei / शाबाश"}
	}
	if strings.Contains(s, "meghalaya") {
		return RegionalProfile{State: state, District: district, DialectName: "KHASI_GARO", ToneHint: "Bha bha eh / Very Good"}
	}
	if strings.Contains(s, "mizoram") {
		return RegionalProfile{State: state, District: district, DialectName: "MIZO", ToneHint: "Tha lutuk / शाबाश"}
	}
	if strings.Contains(s, "nagaland") {
		return RegionalProfile{State: state, District: district, DialectName: "NAGAMESE", ToneHint: "Bhal ase / शाबाश"}
	}
	if strings.Contains(s, "tripura") {
		return RegionalProfile{State: state, District: district, DialectName: "BENGALI_KOKBOROK", ToneHint: "খুব ভালো হয়েছে / शाबाश"}
	}
	if strings.Contains(s, "sikkim") {
		return RegionalProfile{State: state, District: district, DialectName: "NEPALI_SIKKIMESE", ToneHint: "धेरै राम्रो काम / शाबाश"}
	}
	if strings.Contains(s, "arunachal") {
		return RegionalProfile{State: state, District: district, DialectName: "ARUNACHALI_HINDI", ToneHint: "बहुत बढ़िया लिखा है / शाबाश"}
	}

	// ==================== 3. राजस्थान ====================
	if strings.Contains(s, "rajasthan") {
		switch {
		case strings.Contains(d, "ganganagar"), strings.Contains(d, "hanumangarh"), strings.Contains(d, "anupgarh"):
			return RegionalProfile{State: state, District: district, DialectName: "BAGRI", ToneHint: "घणो आछो लिख्यो है / शाबाश टाबर"}
		case strings.Contains(d, "jodhpur"), strings.Contains(d, "bikaner"), strings.Contains(d, "barmer"),
			strings.Contains(d, "nagaur"), strings.Contains(d, "jaisalmer"), strings.Contains(d, "pali"), strings.Contains(d, "jalore"):
			return RegionalProfile{State: state, District: district, DialectName: "MARWARI", ToneHint: "घणी फूटरी राइटिंग / शाबाश बेटा"}
		case strings.Contains(d, "jaipur"), strings.Contains(d, "dausa"), strings.Contains(d, "dudu"), strings.Contains(d, "kotputli"):
			return RegionalProfile{State: state, District: district, DialectName: "DHUNDHARI", ToneHint: "घणो चोखो काम करयो / बहुत बढ़िया"}
		case strings.Contains(d, "sikar"), strings.Contains(d, "jhunjhunu"), strings.Contains(d, "churu"), strings.Contains(d, "neem ka thana"):
			return RegionalProfile{State: state, District: district, DialectName: "SHEKHAWATI", ToneHint: "घणो जोरदार काम / शाबाश लाडले"}
		case strings.Contains(d, "kota"), strings.Contains(d, "bundi"), strings.Contains(d, "baran"), strings.Contains(d, "jhalawar"):
			return RegionalProfile{State: state, District: district, DialectName: "HADOTI", ToneHint: "घणो चोखो बेटा / शाबाश"}
		case strings.Contains(d, "udaipur"), strings.Contains(d, "chittorgarh"), strings.Contains(d, "bhilwara"), strings.Contains(d, "rajsamand"):
			return RegionalProfile{State: state, District: district, DialectName: "MEWARI", ToneHint: "घणो प्यारो लिख्यो / शाबाश"}
		case strings.Contains(d, "dungarpur"), strings.Contains(d, "banswara"), strings.Contains(d, "pratapgarh"):
			return RegionalProfile{State: state, District: district, DialectName: "VAGADI", ToneHint: "खूब हूंदो काम करयो / शाबाश"}
		case strings.Contains(d, "alwar"), strings.Contains(d, "bharatpur"), strings.Contains(d, "deeg"):
			return RegionalProfile{State: state, District: district, DialectName: "MEWATI", ToneHint: "घणा बढ़िया काम कियो / शाबाश"}
		default:
			return RegionalProfile{State: state, District: district, DialectName: "STANDARD_RAJASTHANI", ToneHint: "बहुत बढ़िया / शाबाश"}
		}
	}

	// ==================== 4. उत्तर भारत (हरियाणा, पंजाब, HP, UK) ====================
	if strings.Contains(s, "haryana") {
		return RegionalProfile{State: state, District: district, DialectName: "HARYANVI", ToneHint: "घणा बढ़िया लिख्या लाडले / शाबाश बालक"}
	}
	if strings.Contains(s, "punjab") {
		return RegionalProfile{State: state, District: district, DialectName: "PUNJABI", ToneHint: "ਬਹੁਤ ਵਧੀਆ ਕੰਮ / ਸ਼ਾਬਾਸ਼ ਪੁੱਤਰ"}
	}
	if strings.Contains(s, "himachal") || strings.Contains(s, "uttarakhand") {
		return RegionalProfile{State: state, District: district, DialectName: "PAHADI_GARHWALI", ToneHint: "बहुत बढ़िया काम / शाबाश"}
	}

	// ==================== 5. बिहार व झारखण्ड ====================
	if strings.Contains(s, "bihar") || strings.Contains(s, "jharkhand") {
		switch {
		case strings.Contains(d, "patna"), strings.Contains(d, "gaya"), strings.Contains(d, "bhojpur"), strings.Contains(d, "siwan"), strings.Contains(d, "chhapra"):
			return RegionalProfile{State: state, District: district, DialectName: "BHOJPURI", ToneHint: "बहुत बढ़िया बबुआ / गर्दा उड़ा दिए"}
		case strings.Contains(d, "darbhanga"), strings.Contains(d, "madhubani"), strings.Contains(d, "samastipur"):
			return RegionalProfile{State: state, District: district, DialectName: "MAITHILI", ToneHint: "बहुत नीक लिखलहुँ अछि / शाबाश"}
		case strings.Contains(d, "ranchi"), strings.Contains(d, "dumka"), strings.Contains(d, "dhanbad"):
			return RegionalProfile{State: state, District: district, DialectName: "NAGPURI_SANTHALI", ToneHint: "खूब बेस काम / शाबाश"}
		default:
			return RegionalProfile{State: state, District: district, DialectName: "BIHARI_HINDI", ToneHint: "बहुत सुंदर प्रयास / शाबाश"}
		}
	}

	// ==================== 6. उत्तर प्रदेश ====================
	if strings.Contains(s, "uttar pradesh") || strings.Contains(s, "up") {
		switch {
		case strings.Contains(d, "varanasi"), strings.Contains(d, "gorakhpur"), strings.Contains(d, "ballia"), strings.Contains(d, "azamgarh"):
			return RegionalProfile{State: state, District: district, DialectName: "PURVANCHAL_BHOJPURI", ToneHint: "शानदार लिखले बाड़ा / शाबाश बबुआ"}
		case strings.Contains(d, "lucknow"), strings.Contains(d, "ayodhya"), strings.Contains(d, "kanpur"):
			return RegionalProfile{State: state, District: district, DialectName: "AWADHI", ToneHint: "बहुत सुंदर लिखावट / शाबाश"}
		case strings.Contains(d, "mathura"), strings.Contains(d, "agra"), strings.Contains(d, "aligarh"):
			return RegionalProfile{State: state, District: district, DialectName: "BRAJ_BHASHA", ToneHint: "घणो नीको काम कियौ / शाबाश"}
		default:
			return RegionalProfile{State: state, District: district, DialectName: "STANDARD_HINDI", ToneHint: "बहुत बढ़िया / शाबाश"}
		}
	}

	// ==================== 7. पश्चिम व मध्य भारत (गुजरात, महाराष्ट्र, MP, CG) ====================
	if strings.Contains(s, "gujarat") {
		return RegionalProfile{State: state, District: district, DialectName: "GUJARATI", ToneHint: "ખૂબ સરસ કામ કર્યું / શાબાશ"}
	}
	if strings.Contains(s, "maharashtra") || strings.Contains(s, "goa") {
		return RegionalProfile{State: state, District: district, DialectName: "MARATHI", ToneHint: "खूप छान लिहिलं आहेस / उत्तम"}
	}
	if strings.Contains(s, "madhya pradesh") || strings.Contains(s, "mp") {
		if strings.Contains(d, "indore") || strings.Contains(d, "ujjain") {
			return RegionalProfile{State: state, District: district, DialectName: "MALVI", ToneHint: "घणो आछो काम कियो / शाबाश"}
		}
		return RegionalProfile{State: state, District: district, DialectName: "BUNDELI_HINDI", ToneHint: "बहुत बढ़िया काम / शाबाश"}
	}
	if strings.Contains(s, "chhattisgarh") || strings.Contains(s, "cg") {
		return RegionalProfile{State: state, District: district, DialectName: "CHHATTISGARHI", ToneHint: "गजब बढ़िया लिखे हस / शाबाश टूरा"}
	}

	// ==================== 8. पूर्वी व दक्षिण भारत ====================
	if strings.Contains(s, "bengal") {
		return RegionalProfile{State: state, District: district, DialectName: "BENGALI", ToneHint: "খুব সুন্দর লিখেছো / দারুণ"}
	}
	if strings.Contains(s, "odisha") || strings.Contains(s, "orissa") {
		return RegionalProfile{State: state, District: district, DialectName: "ODIA", ToneHint: "ବହୁତ ଭଲ ଲେଖିଛ / ଶାବାସ"}
	}
	if strings.Contains(s, "telangana") || strings.Contains(s, "andhra") {
		return RegionalProfile{State: state, District: district, DialectName: "TELUGU", ToneHint: "చాలా బాగుంది / శభాష్"}
	}
	if strings.Contains(s, "tamil") {
		return RegionalProfile{State: state, District: district, DialectName: "TAMIL", ToneHint: "மிகவும் அருமை / நன்று"}
	}
	if strings.Contains(s, "karnataka") {
		return RegionalProfile{State: state, District: district, DialectName: "KANNADA", ToneHint: "ತುಂಬಾ ಚೆನ್ನಾಗಿದೆ / ಶಾಬಾಶ್"}
	}
	if strings.Contains(s, "kerala") {
		return RegionalProfile{State: state, District: district, DialectName: "MALAYALAM", ToneHint: "വളരെ നന്നായിരിക്കുന്നു / മിടുക്കൻ"}
	}

	// डिफ़ॉल्ट: अखिल भारतीय मानक हिंदी
	return RegionalProfile{
		State:       state,
		District:    district,
		DialectName: "STANDARD_HINDI",
		ToneHint:    "बहुत बढ़िया लिखा है / शाबाश",
	}
}
