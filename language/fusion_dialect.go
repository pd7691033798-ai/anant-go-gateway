package language

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type DialectProfile struct {
	DialectCode string `json:"dialect_code"`
	ToneHint    string `json:"tone_hint"`
	RegionHint  string `json:"region_hint"`
	IsFusion    bool   `json:"is_fusion"`
}

type FusionDialectService struct {
	db *sql.DB
}

func NewFusionDialectService(db *sql.DB) *FusionDialectService {
	return &FusionDialectService{db: db}
}

// DetectAndResolve छात्र के इनपुट, जिले और सेव्ड प्रोफाइल के आधार पर सटीक बोली तय करता है
func (f *FusionDialectService) DetectAndResolve(ctx context.Context, studentUID, rawInput, state, district string) DialectProfile {
	// 1. यदि डेटाबेस में छात्र की प्राथमिकता पहले से सेव है, तो उसे चेक करें
	if f.db != nil && studentUID != "" {
		var savedDialect string
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		query := `SELECT COALESCE(preferred_dialect, '') FROM onboarding_sessions WHERE phone = $1 OR phone = (SELECT phone FROM users WHERE uid_badge = $1 LIMIT 1)`
		if err := f.db.QueryRowContext(dbCtx, query, studentUID).Scan(&savedDialect); err == nil && savedDialect != "" {
			switch savedDialect {
			case "RAJASTHANI_BAGRI":
				return DialectProfile{DialectCode: "BAGRI", ToneHint: "घणो आछो लिख्यो है / शाबाश टाबर", RegionHint: "उत्तर/पश्चिम राजस्थान", IsFusion: false}
			case "HARYANVI":
				return DialectProfile{DialectCode: "HARYANVI", ToneHint: "घणा बढ़िया काम करया लाडले!", RegionHint: "हरियाणा", IsFusion: false}
			case "BHOJPURI":
				return DialectProfile{DialectCode: "BHOJPURI", ToneHint: "बहुत बढ़िया बबुआ, आगे बढ़ा जाए!", RegionHint: "बिहार / पूर्वी यूपी", IsFusion: false}
			case "PUNJABI":
				return DialectProfile{DialectCode: "PUNJABI", ToneHint: "ਬਹੁਤ ਵਧੀਆ ਪੁੱਤਰ! ਚੱਕ ਦੇ ਫੱਟੇ!", RegionHint: "पंजाब", IsFusion: false}
			}
		}
	}

	// 2. इनपुट टेक्स्ट का विश्लेषण (देवनागरी + हिंग्लिश/रोमन दोनों के लिए)
	lower := strings.ToLower(rawInput)
	cleanDistrict := strings.ToLower(strings.TrimSpace(district))
	cleanState := strings.ToLower(strings.TrimSpace(state))

	// बागड़ी/राजस्थानी कीवर्ड्स
	hasBagri := strings.Contains(lower, "ठा कोनी") || strings.Contains(lower, "था कोनी") ||
		strings.Contains(lower, "म्हारो") || strings.Contains(lower, "घणो") ||
		strings.Contains(lower, "टाबर") || strings.Contains(lower, "tha koni")

	// पंजाबी कीवर्ड्स
	hasPunjabi := strings.Contains(lower, "दस") || strings.Contains(lower, "चाहना") ||
		strings.Contains(lower, "ਕੀ") || strings.Contains(lower, "ਪੁੱਤਰ") ||
		strings.Contains(lower, "ਚੱਕ") || strings.Contains(lower, "kidaan") ||
		strings.Contains(lower, "veere") || strings.Contains(lower, "tussi")

	// हरियाणवी कीवर्ड्स
	hasHaryanvi := strings.Contains(lower, "बेरा कोनी") || strings.Contains(lower, "लाडले") ||
		strings.Contains(lower, "मन्ने") || strings.Contains(lower, "तन्ने") ||
		strings.Contains(lower, "bera koni")

	// 3. फ़्यूज़न डिटेक्शन (सीमावर्ती क्षेत्र जैसे श्रीगंगानगर, अबोहर, फाजिल्का, सिरसा)
	if (hasBagri && hasPunjabi) || (strings.Contains(cleanDistrict, "ganganagar") && hasPunjabi) {
		return DialectProfile{
			DialectCode: "BAGRI_PUNJABI_FUSION",
			ToneHint:    "मन्ने सब ठा है, तू की कहना चौहना ऐ! चल कॉपी चक्क ते लिख!",
			RegionHint:  "श्रीगंगानगर/फाजिल्का सीमावर्ती बेल्ट",
			IsFusion:    true,
		}
	}

	// 4. ज़िला और राज्य आधारित मैपिंग
	if strings.Contains(cleanDistrict, "ganganagar") || strings.Contains(cleanDistrict, "hanumangarh") || hasBagri {
		return DialectProfile{DialectCode: "BAGRI", ToneHint: "घणो आछो लिख्यो है / शाबाश टाबर", RegionHint: "उत्तर राजस्थान", IsFusion: false}
	}

	if strings.Contains(cleanState, "haryana") || strings.Contains(cleanDistrict, "sirsa") || strings.Contains(cleanDistrict, "hisar") || hasHaryanvi {
		return DialectProfile{DialectCode: "HARYANVI", ToneHint: "घणा बढ़िया काम करया लाडले!", RegionHint: "हरियाणा", IsFusion: false}
	}

	if hasPunjabi || strings.Contains(cleanState, "punjab") {
		return DialectProfile{DialectCode: "PUNJABI", ToneHint: "ਬਹੁਤ ਵਧੀਆ ਪੁੱਤਰ! ਚੱਕ ਦੇ ਫੱਟੇ!", RegionHint: "पंजाब", IsFusion: false}
	}

	// डिफ़ॉल्ट मानक हिंदी
	return DialectProfile{
		DialectCode: "STANDARD_HINDI",
		ToneHint:    "बहुत बढ़िया लिखा है / शाबाश",
		RegionHint:  "अखिल भारतीय",
		IsFusion:    false,
	}
}
