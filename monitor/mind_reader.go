package monitor

import (
	"fmt"
	"strings"
)

type MindReader struct{}

func NewMindReader() *MindReader {
	return &MindReader{}
}

// ProbeCheekyStudent नटखट या टालमटोल भरे जवाबों को पहचानकर बच्चे को दोस्ताना अंदाज़ में पढ़ाई पर लाता है
func (m *MindReader) ProbeCheekyStudent(studentName, message, dialect string) (bool, string) {
	name := strings.TrimSpace(studentName)
	if name == "" {
		name = "दोस्त"
	}

	lower := strings.ToLower(strings.TrimSpace(message))
	d := strings.ToUpper(strings.TrimSpace(dialect))

	// 1. नटखट / टालमटोल भरे कीवर्ड्स (हिंदी + हिंग्लिश + क्षेत्रीय बोलियाँ)
	isCheeky := strings.Contains(lower, "क्यों बताऊँ") || strings.Contains(lower, "क्यों बताऊं") ||
		strings.Contains(lower, "तुम जानो") || strings.Contains(lower, "नहीं बताऊंगा") ||
		strings.Contains(lower, "नहीं बताउंगा") || strings.Contains(lower, "kyu batau") ||
		strings.Contains(lower, "kyun bataun") || strings.Contains(lower, "nahi bataunga") ||
		strings.Contains(lower, "tumhe kya") || strings.Contains(lower, "apne aap dekh lo") ||
		strings.Contains(lower, "कोनी बताऊँ") || strings.Contains(lower, "koni batau") ||
		strings.Contains(lower, "ना बताइब") || strings.Contains(lower, "na bataib") ||
		strings.Contains(lower, "main kyu dassa") || strings.Contains(lower, "ਕਿਉਂ ਦੱਸਾਂ")

	if !isCheeky {
		return false, "बहुत बढ़िया! चलिए आज के 15-मिनट अभ्यास पर ध्यान देते हैं।"
	}

	// 2. बोली के अनुसार मज़ेदार व प्रेरक जवाब
	switch d {
	case "BAGRI", "MARWARI":
		return true, fmt.Sprintf("अरे %s! 🕵️‍♂️ आ होशियारी छोड़ो अर चुपचाप आज को 15-मिनट को अभ्यास पूरा करो, ताकी मेडल पक्को रहे!", name)

	case "HARYANVI":
		return true, fmt.Sprintf("अरे %s लाडले! 🕵️‍♂️ जासूसी मूड छोड़ अर आज का 15-मिनट का अभ्यास कर ले, ना तै स्ट्रीक टूट ज्यागी!", name)

	case "BHOJPURI":
		return true, fmt.Sprintf("अरे %s बबुआ! 🕵️‍♂️ नौटंकी बंद करा आ जल्दी से 15-मिनट के अभ्यास पूरा करा, मेडल बचावे के बा!", name)

	case "PUNJABI":
		return true, fmt.Sprintf("ਅਰੇ %s ਪੁੱਤਰ! 🕵️‍♂️ ਚਲਾਕੀਆਂ ਛੱਡੋ ਤੇ ਅੱਜ ਦਾ 15-ਮਿੰਟ ਅਭਿਆਸ ਪੂਰਾ ਕਰੋ ਤਾਂ ਕਿ ਮੈਡਲ ਪੱਕਾ ਰਹੇ!", name)

	default:
		return true, fmt.Sprintf("अरे %s! 🕵️‍♂️ जासूसी मूड छोड़िए और आज का 15-मिनट अभ्यास भेजिए ताकि आपका मेडल सुरक्षित रहे!", name)
	}
}

