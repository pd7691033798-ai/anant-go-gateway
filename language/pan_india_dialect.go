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

type PanIndiaDialectService struct{}

func NewPanIndiaDialectService() *PanIndiaDialectService {
	return &PanIndiaDialectService{}
}

// cleanString स्पेस हटाकर लोअरकेस बनाता है ताकि मैचिंग कभी फेल न हो
func cleanString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (p *PanIndiaDialectService) ResolveRegionalDialect(state, district string) RegionalProfile {
	s := cleanString(state)
	d := cleanString(district)

	// ==================== 1. राजस्थान (Rajasthan) ====================
	if strings.Contains(s, "rajasthan") {
		switch {
		// उत्तर राजस्थान (बागड़ी बेल्ट)
		case strings.Contains(d, "ganganagar"), strings.Contains(d, "hanumangarh"), strings.Contains(d, "anupgarh"):
			return RegionalProfile{State: state, District: district, DialectName: "BAGRI", ToneHint: "घणो आछो लिख्यो है / शाबाश टाबर"}

		// मारवाड़ बेल्ट
		case strings.Contains(d, "jodhpur"), strings.Contains(d, "bikaner"), strings.Contains(d, "barmer"),
			strings.Contains(d, "nagaur"), strings.Contains(d, "jaisalmer"), strings.Contains(d, "pali"), strings.Contains(d, "jalore"):
			return RegionalProfile{State: state, District: district, DialectName: "MARWARI", ToneHint: "घणी फूटरी राइटिंग / शाबाश बेटा"}

		// ढूंढाड़ बेल्ट
		case strings.Contains(d, "jaipur"), strings.Contains(d, "dausa"), strings.Contains(d, "dudu"), strings.Contains(d, "kotputli"):
			return RegionalProfile{State: state, District: district, DialectName: "DHUNDHARI", ToneHint: "घणो चोखो काम करयो / बहुत बढ़िया"}

		// शेखावाटी बेल्ट
		case strings.Contains(d, "sikar"), strings.Contains(d, "jhunjhunu"), strings.Contains(d, "churu"), strings.Contains(d, "neem ka thana"):
			return RegionalProfile{State: state, District: district, DialectName: "SHEKHAWATI", ToneHint: "घणो जोरदार काम / शाबाश लाडले"}

		// हाड़ौती बेल्ट
		case strings.Contains(d, "kota"), strings.Contains(d, "bundi"), strings.Contains(d, "baran"), strings.Contains(d, "jhalawar"):
			return RegionalProfile{State: state, District: district, DialectName: "HADOTI", ToneHint: "घणो चोखो बेटा / शाबाश"}

		// मेवाड़ बेल्ट
		case strings.Contains(d, "udaipur"), strings.Contains(d, "chittorgarh"), strings.Contains(d, "bhilwara"), strings.Contains(d, "rajsamand"):
			return RegionalProfile{State: state, District: district, DialectName: "MEWARI", ToneHint: "घणो प्यारो लिख्यो / शाबाश"}

		// वागड़ बेल्ट
		case strings.Contains(d, "dungarpur"), strings.Contains(d, "banswara"), strings.Contains(d, "pratapgarh"):
			return RegionalProfile{State: state, District: district, DialectName: "VAGADI", ToneHint: "खूब हूंदो काम करयो / शाबाश"}

		// मेवात बेल्ट
		case strings.Contains(d, "alwar"), strings.Contains(d, "bharatpur"), strings.Contains(d, "deeg"):
			return RegionalProfile{State: state, District: district, DialectName: "MEWATI", ToneHint: "घणा बढ़िया काम कियो / शाबाश"}

		default:
			return RegionalProfile{State: state, District: district, DialectName: "STANDARD_RAJASTHANI", ToneHint: "बहुत बढ़िया / शाबाश"}
		}
	}

	// ==================== 2. हरियाणा (Haryana) ====================
	if strings.Contains(s, "haryana") {
		return RegionalProfile{State: state, District: district, DialectName: "HARYANVI", ToneHint: "घणा बढ़िया लिख्या लाडले / शाबाश बालक"}
	}

	// ==================== 3. पंजाब (Punjab) ====================
	if strings.Contains(s, "punjab") {
		return RegionalProfile{State: state, District: district, DialectName: "PUNJABI", ToneHint: "ਬਹੁਤ ਵਧੀਆ ਕੰਮ / ਸ਼ਾਬਾਸ਼ ਪੁੱਤਰ"}
	}

	// ==================== 4. बिहार (Bihar) ====================
	if strings.Contains(s, "bihar") {
		switch {
		case strings.Contains(d, "patna"), strings.Contains(d, "gaya"), strings.Contains(d, "bhojpur"), strings.Contains(d, "siwan"), strings.Contains(d, "chhapra"):
			return RegionalProfile{State: state, District: district, DialectName: "BHOJPURI", ToneHint: "बहुत बढ़िया बबुआ / गर्दा उड़ा दिए"}
		case strings.Contains(d, "darbhanga"), strings.Contains(d, "madhubani"), strings.Contains(d, "samastipur"):
			return RegionalProfile{State: state, District: district, DialectName: "MAITHILI", ToneHint: "बहुत नीक लिखलहुँ अछि / शाबाश"}
		default:
			return RegionalProfile{State: state, District: district, DialectName: "BIHARI_HINDI", ToneHint: "बहुत सुंदर प्रयास / शाबाश"}
		}
	}

	// ==================== 5. उत्तर प्रदेश (Uttar Pradesh) ====================
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

	// ==================== 6. गुजरात (Gujarat) ====================
	if strings.Contains(s, "gujarat") {
		return RegionalProfile{State: state, District: district, DialectName: "GUJARATI", ToneHint: "ખૂબ સરસ કામ કર્યું / શાબાશ"}
	}

	// ==================== 7. महाराष्ट्र (Maharashtra) ====================
	if strings.Contains(s, "maharashtra") {
		return RegionalProfile{State: state, District: district, DialectName: "MARATHI", ToneHint: "खूप छान लिहिलं आहेस / उत्तम"}
	}

	// ==================== 8. मध्य प्रदेश व छत्तीसगढ़ (MP & CG) ====================
	if strings.Contains(s, "madhya pradesh") || strings.Contains(s, "mp") {
		if strings.Contains(d, "indore") || strings.Contains(d, "ujjain") {
			return RegionalProfile{State: state, District: district, DialectName: "MALVI", ToneHint: "घणो आछो काम कियो / शाबाश"}
		}
		return RegionalProfile{State: state, District: district, DialectName: "BUNDELI_HINDI", ToneHint: "बहुत बढ़िया काम / शाबाश"}
	}
	if strings.Contains(s, "chhattisgarh") || strings.Contains(s, "cg") {
		return RegionalProfile{State: state, District: district, DialectName: "CHHATTISGARHI", ToneHint: "गजब बढ़िया लिखे हस / शाबाश टूरा"}
	}

	// ==================== 9. पश्चिम बंगाल (West Bengal) ====================
	if strings.Contains(s, "bengal") {
		return RegionalProfile{State: state, District: district, DialectName: "BENGALI", ToneHint: "খুব সুন্দর লিখেছো / দারুণ"}
	}

	// ==================== 10. दक्षिण भारत (South India) ====================
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

	// ==================== 11. ओडिशा व पूर्वोत्तर ====================
	if strings.Contains(s, "odisha") || strings.Contains(s, "orissa") {
		return RegionalProfile{State: state, District: district, DialectName: "ODIA", ToneHint: "ବହୁତ ଭଲ ଲେଖିଛ / ଶାବାସ"}
	}
	if strings.Contains(s, "assam") {
		return RegionalProfile{State: state, District: district, DialectName: "ASSAMESE", ToneHint: "বৰ ধুনীয়া হৈছে / শাবাশ"}
	}

	// ==================== 12. पहाड़ी क्षेत्र (HP & Uttarakhand) ====================
	if strings.Contains(s, "himachal") || strings.Contains(s, "uttarakhand") {
		return RegionalProfile{State: state, District: district, DialectName: "PAHADI_GARHWALI", ToneHint: "बहुत बढ़िया काम / शाबाश"}
	}

	// डिफ़ॉल्ट: अखिल भारतीय मानक हिंदी
	return RegionalProfile{
		State:       state,
		District:    district,
		DialectName: "STANDARD_HINDI",
		ToneHint:    "बहुत बढ़िया लिखा है / शाबाश",
	}
}
