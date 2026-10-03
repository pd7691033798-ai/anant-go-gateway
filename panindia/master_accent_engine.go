package panindia

import (
	"strings"
)

// AccentProfile भाषा, क्षेत्रीय एक्सेंट और मानव जैसी बोलने की गति (Speech Cadence) को संभालती है
type AccentProfile struct {
	StateName       string   `json:"state_name"`
	LanguageName    string   `json:"language_name"`
	DefaultAccent   string   `json:"default_accent"`
	RegionalAccents []string `json:"regional_accents"`
	SpeechCadence   string   `json:"speech_cadence"`   // बोलने की गति और प्राकृतिक अंदाज़ (जैसे: Rapid, Steady, Melodic)
	ToneStyle       string   `json:"tone_style"`       // बातचीत का मानवीय लहजा (जैसे: Warm, Energetic, Calm)
}

type MasterPanIndiaEngine struct {
	DefaultLanguage string
	DefaultAccent   string
	DefaultCadence  string
}

func NewMasterPanIndiaEngine() *MasterPanIndiaEngine {
	return &MasterPanIndiaEngine{
		DefaultLanguage: "Hindi",
		DefaultAccent:   "Standard Khadi Boli",
		DefaultCadence:  "Natural conversational pace with warm human pauses",
	}
}

// StateLanguageMap: भारत के सभी 28 राज्यों की भाषाएं, एक्सेंट और उनके मानवीय बोलने के तरीके की मास्टर डिक्शनरी
var StateLanguageMap = map[string]AccentProfile{
	"andhra pradesh": {
		StateName: "Andhra Pradesh", LanguageName: "Telugu", DefaultAccent: "Coastal Telugu",
		RegionalAccents: []string{"Coastal Telugu", "Rayalaseema Accent", "North Coastal Accent"},
		SpeechCadence:   "Melodic flow with expressive syllabic stress and natural conversational pauses",
		ToneStyle:       "Warm, respectful, and engaging teacher tone",
	},
	"arunachal pradesh": {
		StateName: "Arunachal Pradesh", LanguageName: "English / Local Dialects", DefaultAccent: "Standard English",
		RegionalAccents: []string{"Adi", "Nishi", "Monpa", "Tani"},
		SpeechCadence:   "Calm, measured cadence with soft articulation",
		ToneStyle:       "Friendly, polite, and welcoming",
	},
	"assam": {
		StateName: "Assam", LanguageName: "Assamese", DefaultAccent: "Central Assamese",
		RegionalAccents: []string{"Kamrupi", "Goalpariya", "Upper Assamese", "Barak Valley Bengali"},
		SpeechCadence:   "Gentle, unhurried, and soft-spoken conversational flow",
		ToneStyle:       "Calm, nurturing, and helpful",
	},
	"bihar": {
		StateName: "Bihar", LanguageName: "Hindi / Bhojpuri / Maithili", DefaultAccent: "Bhojpuri / Magahi",
		RegionalAccents: []string{"Bhojpuri", "Maithili", "Magahi", "Angika", "Vajjika"},
		SpeechCadence:   "Lively, expressive pacing with strong emotional connect and friendly warmth",
		ToneStyle:       "Energetic, down-to-earth, and brotherly/supportive",
	},
	"chhattisgarh": {
		StateName: "Chhattisgarh", LanguageName: "Chhattisgarhi / Hindi", DefaultAccent: "Standard Chhattisgarhi",
		RegionalAccents: []string{"Standard Chhattisgarhi", "Sargujia", "Halbi"},
		SpeechCadence:   "Rhythmic and steady pacing with a grounded rural warmth",
		ToneStyle:       "Simple, direct, and affectionate",
	},
	"goa": {
		StateName: "Goa", LanguageName: "Konkani / Marathi", DefaultAccent: "Antruz Konkani",
		RegionalAccents: []string{"Antruz", "Sashti", "Bardez", "Malvani/Goan Konkani"},
		SpeechCadence:   "Relaxed, breezy, and melodious conversational rhythm",
		ToneStyle:       "Cheerful, casual, and encouraging",
	},
	"gujarat": {
		StateName: "Gujarat", LanguageName: "Gujarati", DefaultAccent: "Standard Gujarati",
		RegionalAccents: []string{"Kathiyawadi", "Charotari", "Surati", "Kutchhi", "Patani"},
		SpeechCadence:   "Quick, crisp, and articulate flow with an upbeat business-like rhythm",
		ToneStyle:       "Enthusiastic, sharp, and motivating",
	},
	"haryana": {
		StateName: "Haryana", LanguageName: "Haryanvi / Hindi", DefaultAccent: "Bangru",
		RegionalAccents: []string{"Bangru", "Deshwali", "Ahirwati", "Bagri"},
		SpeechCadence:   "Bold, high-energy, and punchy delivery with strong emphasis",
		ToneStyle:       "Direct, fearless, and high-spirited",
	},
	"himachal pradesh": {
		StateName: "Himachal Pradesh", LanguageName: "Pahadi / Hindi", DefaultAccent: "Standard Pahadi",
		RegionalAccents: []string{"Kangri", "Mandeali", "Kulvi", "Chambeali"},
		SpeechCadence:   "Calm, serene, and unhurried natural pace",
		ToneStyle:       "Soothing, polite, and patient",
	},
	"jharkhand": {
		StateName: "Jharkhand", LanguageName: "Hindi / Khortha / Santhali", DefaultAccent: "Khortha / Sadri",
		RegionalAccents: []string{"Khortha", "Nagpuri", "Kurmali", "Panchpargania", "Santhali"},
		SpeechCadence:   "Grounded, steady, and warm pacing",
		ToneStyle:       "Sincere, earnest, and supportive",
	},
	"karnataka": {
		StateName: "Karnataka", LanguageName: "Kannada", DefaultAccent: "Mysuru Kannada",
		RegionalAccents: []string{"Mysuru / Bengaluru", "Dharwad / North Karnataka", "Mangaluru / Coastal", "Kundapura"},
		SpeechCadence:   "Fluid, rhythmic, and clear pronunciation with natural conversational pauses",
		ToneStyle:       "Polite, intelligent, and encouraging",
	},
	"kerala": {
		StateName: "Kerala", LanguageName: "Malayalam", DefaultAccent: "Central Kerala Malayalam",
		RegionalAccents: []string{"Central Kerala", "Malabar (North)", "Travancore (South)", "Kochi Accent"},
		SpeechCadence:   "Smooth, articulate, and moderately fast glide with distinct vowel clarity",
		ToneStyle:       "Intellectual, courteous, and precise",
	},
	"madhya pradesh": {
		StateName: "Madhya Pradesh", LanguageName: "Hindi", DefaultAccent: "Standard Hindi",
		RegionalAccents: []string{"Malvi", "Nimadi", "Bundeli", "Bagheli", "Chhattisgarhi mix"},
		SpeechCadence:   "Balanced, easygoing, and natural conversational cadence",
		ToneStyle:       "Friendly, relaxed, and helpful",
	},
	"maharashtra": {
		StateName: "Maharashtra", LanguageName: "Marathi", DefaultAccent: "Puneri / Standard Marathi",
		RegionalAccents: []string{"Varhadi (Vidarbha)", "Konkani / Malvani", "Ahirani (Khandesh)", "Deshi / Marathwada", "Puneri / Mumbai Standard"},
		SpeechCadence:   "Crisp, sharp, and rhythmic pacing; dynamic modulation when switching to regional dialects like Varhadi",
		ToneStyle:       "Smart, confident, and articulate mentor tone",
	},
	"manipur": {
		StateName: "Manipur", LanguageName: "Meitei (Manipuri)", DefaultAccent: "Standard Meiteilon",
		RegionalAccents: []string{"Imphal Standard", "Bishnupur dialect", "Thoubal dialect"},
		SpeechCadence:   "Melodic, soft-toned, and steady rhythmic flow",
		ToneStyle:       "Graceful, respectful, and calm",
	},
	"meghalaya": {
		StateName: "Meghalaya", LanguageName: "Khasi / Garo", DefaultAccent: "Standard Khasi",
		RegionalAccents: []string{"Sohra Khasi", "Pnar", "Garo (A'we, Matabeng)"},
		SpeechCadence:   "Soft, deliberate, and clear conversational pace",
		ToneStyle:       "Warm and inviting",
	},
	"mizoram": {
		StateName: "Mizoram", LanguageName: "Mizo", DefaultAccent: "Lusei (Duhlian)",
		RegionalAccents: []string{"Lusei Standard", "Hmar", "Mara", "Lai"},
		SpeechCadence:   "Harmonious and flowing tempo",
		ToneStyle:       "Kind, communal, and supportive",
	},
	"nagaland": {
		StateName: "Nagaland", LanguageName: "Nagamese / English", DefaultAccent: "Nagamese Creole",
		RegionalAccents: []string{"Nagamese", "Ao Naga", "Angami", "Sema", "Konyak"},
		SpeechCadence:   "Lively, crisp, and direct conversational style",
		ToneStyle:       "Youthful, energetic, and casual",
	},
	"odisha": {
		StateName: "Odisha", LanguageName: "Odia", DefaultAccent: "Coastal Odia",
		RegionalAccents: []string{"Coastal Odia", "Western Odia (Sambalpuri)", "Northern Odia", "Desia"},
		SpeechCadence:   "Sweet, soft-inflected, and steady rhythmic flow",
		ToneStyle:       "Humble, caring, and educational",
	},
	"punjab": {
		StateName: "Punjab", LanguageName: "Punjabi", DefaultAccent: "Majhi",
		RegionalAccents: []string{"Majhi (Amritsar/Lahore)", "Malwai", "Doabi", "Pwadhi"},
		SpeechCadence:   "Vibrant, high-energy, robust cadence with powerful emotional resonance",
		ToneStyle:       "Passionate, large-hearted, and deeply encouraging",
	},
	"rajasthan": {
		StateName: "Rajasthan", LanguageName: "Rajasthani / Hindi", DefaultAccent: "Marwari / Dhundhari",
		RegionalAccents: []string{"Marwari", "Mewari", "Dhundhari (Jaipur)", "Shekhawati", "Bagri", "Malvi"},
		SpeechCadence:   "Majestic, warm, and flowing conversational rhythm with royal courtesy",
		ToneStyle:       "Respectful, hospitable, and inspiring",
	},
	"sikkim": {
		StateName: "Sikkim", LanguageName: "Nepali / Bhutia / Lepcha", DefaultAccent: "Standard Nepali",
		RegionalAccents: []string{"Gorkha Nepali", "Bhutia", "Lepcha"},
		SpeechCadence:   "Clear, melodic, and pleasant pacing",
		ToneStyle:       "Peaceful, polite, and warm",
	},
	"tamil nadu": {
		StateName: "Tamil Nadu", LanguageName: "Tamil", DefaultAccent: "Chennai / Standard Tamil",
		RegionalAccents: []string{"Kongu Tamil", "Madurai / Pandiyan", "Tirunelveli Accent", "Chennai Tamil", "Sri Lankan Tamil mix"},
		SpeechCadence:   "Super-fast, agile, dynamic rapid-fire delivery with natural agglutinative word bonding and high energy",
		ToneStyle:       "Sharp, fast-witted, articulate, and deeply passionate",
	},
	"telangana": {
		StateName: "Telangana", LanguageName: "Telugu", DefaultAccent: "Telangana Telugu",
		RegionalAccents: []string{"Telangana Dialect (Deccan influence)", "Hyderabad Urdu-Telugu mix"},
		SpeechCadence:   "Expressive, punchy, folk-driven rhythm with high conversational zest",
		ToneStyle:       "Vibrant, bold, and engaging",
	},
	"tripura": {
		StateName: "Tripura", LanguageName: "Bengali / Kokborok", DefaultAccent: "Tripuri Bengali",
		RegionalAccents: []string{"Sylheti/Tripura Bengali", "Kokborok (Tripuri)"},
		SpeechCadence:   "Fluid and expressive storytelling cadence",
		ToneStyle:       "Warm and welcoming",
	},
	"uttar pradesh": {
		StateName: "Uttar Pradesh", LanguageName: "Hindi", DefaultAccent: "Khadi Boli",
		RegionalAccents: []string{"Awadhi", "Braj Bhasha", "Bhojpuri (Purvanchal)", "Khadi Boli", "Kannauji", "Rohilkhandi"},
		SpeechCadence:   "Expressive, storytelling pace with natural poetic pauses (adapting to Awadhi/Braj sweetness or Khadi clarity)",
		ToneStyle:       "Engaging, wise, and brotherly mentor tone",
	},
	"uttarakhand": {
		StateName: "Uttarakhand", LanguageName: "Garhwali / Kumaoni / Hindi", DefaultAccent: "Garhwali / Kumaoni",
		RegionalAccents: []string{"Garhwali", "Kumaoni", "Jaunsari"},
		SpeechCadence:   "Clear, mountain-fresh, steady pacing",
		ToneStyle:       "Sincere, pure, and encouraging",
	},
	"west bengal": {
		StateName: "West Bengal", LanguageName: "Bengali", DefaultAccent: "Rarhi (Kolkata Standard)",
		RegionalAccents: []string{"Kolkata Standard (Rarhi)", "Bangal (East Bengal/Dhaka mix)", "Varendra (North Bengal)", "Rajbanshi"},
		SpeechCadence:   "Melodious, lyrical, soft-flowing cadence with artistic modulation and artistic vowel draws",
		ToneStyle:       "Intellectual, poetic, warm, and deeply affectionate",
	},
}

// FullVoiceProfile 
// भाषा, एक्सेंट, राज्य के साथ-साथ वॉयस डिलीवरी / स्पीच पेसिंग की पूरी जानकारी देता है
type FullVoiceProfile struct {
	LanguageName  string
	Accent        string
	StateName     string
	SpeechCadence string
	ToneStyle     string
}

// 1. जब कोई फाइल या सिस्टम खुद चलकर मांगे (Explicit Pull Request with Voice Metadata)
func (m *MasterPanIndiaEngine) DetectVoiceAndAccentProfile(inputQuery string) FullVoiceProfile {
	cleaned := strings.ToLower(strings.TrimSpace(inputQuery))

	if cleaned == "" {
		return FullVoiceProfile{
			LanguageName:  m.DefaultLanguage,
			Accent:        m.DefaultAccent,
			StateName:     "All India Default",
			SpeechCadence: m.DefaultCadence,
			ToneStyle:     "Natural human conversational tone",
		}
	}

	// 28 राज्यों और भाषाओं का मिलान करते हुए उनका वॉयस प्रोफाइल निकालना
	for stateKey, profile := range StateLanguageMap {
		if strings.Contains(cleaned, stateKey) || strings.Contains(cleaned, strings.ToLower(profile.LanguageName)) {
			selectedAccent := profile.DefaultAccent
			for _, accent := range profile.RegionalAccents {
				if strings.Contains(cleaned, strings.ToLower(accent)) {
					selectedAccent = accent
					break
				}
			}
			return FullVoiceProfile{
				LanguageName:  profile.LanguageName,
				Accent:        selectedAccent,
				StateName:     profile.StateName,
				SpeechCadence: profile.SpeechCadence,
				ToneStyle:     profile.ToneStyle,
			}
		}
	}

	// विशेष बोलियों (जैसे तमिल, मराठी, भोजपुरी, बंगाली) के लिए डायरेक्ट मैच और उनका खास स्पीच पेसिंग
	if strings.Contains(cleaned, "tamil") || strings.Contains(cleaned, "தமிழ்") {
		p := StateLanguageMap["tamil nadu"]
		return FullVoiceProfile{
			LanguageName: p.LanguageName, Accent: p.DefaultAccent, StateName: p.StateName,
			SpeechCadence: p.SpeechCadence, ToneStyle: p.ToneStyle,
		}
	}
	if strings.Contains(cleaned, "marathi") || strings.Contains(cleaned, "मराठी") {
		p := StateLanguageMap["maharashtra"]
		accent := p.DefaultAccent
		if strings.Contains(cleaned, "varhadi") || strings.Contains(cleaned, "वरहाड़ी") {
			accent = "Varhadi (Vidarbha)"
		}
		return FullVoiceProfile{
			LanguageName: p.LanguageName, Accent: accent, StateName: p.StateName,
			SpeechCadence: p.SpeechCadence, ToneStyle: p.ToneStyle,
		}
	}
	if strings.Contains(cleaned, "bengali") || strings.Contains(cleaned, "বাংলা") {
		p := StateLanguageMap["west bengal"]
		return FullVoiceProfile{
			LanguageName: p.LanguageName, Accent: p.DefaultAccent, StateName: p.StateName,
			SpeechCadence: p.SpeechCadence, ToneStyle: p.ToneStyle,
		}
	}

	// फॉलबैक डिफ़ॉल्ट
	return FullVoiceProfile{
		LanguageName:  m.DefaultLanguage,
		Accent:        m.DefaultAccent,
		StateName:     "Pan-India Fallback",
		SpeechCadence: m.DefaultCadence,
		ToneStyle:     "Warm human conversational tone",
	}
}

// 2. जब इंजन खुद ऑटो-चेक करे (Automatic Inspection & Fallback Logic)
// यह चेक करता है कि क्या किसी फाइल/मॉड्यूल के पास पहले से अपनी वॉयस/लैंग्वेज सेटिंग है या नहीं।
func (m *MasterPanIndiaEngine) AutoCheckAndInspectVoiceModule(existingLang string, existingAccent string, existingCadence string) (string, string, string) {
	finalLang := existingLang
	finalAccent := existingAccent
	finalCadence := existingCadence

	if strings.TrimSpace(finalLang) == "" || finalLang == "NEW" || finalLang == "DEFAULT" {
		finalLang = m.DefaultLanguage
	}

	if strings.TrimSpace(finalAccent) == "" || finalAccent == "NEW" {
		finalAccent = m.DefaultAccent
	}

	if strings.TrimSpace(finalCadence) == "" || finalCadence == "NEW" {
		finalCadence = m.DefaultCadence
	}

	return finalLang, finalAccent, finalCadence
}
