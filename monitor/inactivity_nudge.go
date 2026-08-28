package monitor

import (
	"fmt"
	"strings"
)

type InactivityNudgeService struct{}

func NewInactivityNudgeService() *InactivityNudgeService {
	return &InactivityNudgeService{}
}

// GenerateDynamicNudge छात्र के नाम, छूटे हुए दिनों, वास्तविक स्ट्रीक और बोली के आधार पर सटीक रिमाइंडर बनाता है
func (i *InactivityNudgeService) GenerateDynamicNudge(studentName string, missedDays, currentStreak int, dialect string) string {
	name := strings.TrimSpace(studentName)
	if name == "" {
		name = "बेटा"
	}

	dialect = strings.ToUpper(strings.TrimSpace(dialect))

	switch {
	case missedDays <= 0:
		return "" // कोई रिमायंडर की आवश्यकता नहीं

	case missedDays == 1:
		// दिन 1: हल्का व दोस्ताना स्मरण
		switch dialect {
		case "BAGRI", "MARWARI":
			return fmt.Sprintf("⏰ %s बेटा, आज को 15-मिनट को अभ्यास बाकी है! जल्दी शुरू करो।", name)
		case "HARYANVI":
			return fmt.Sprintf("⏰ %s लाडले, आज का 15-मिनट का अभ्यास रह रह्या सै! जल्दी करो।", name)
		case "BHOJPURI":
			return fmt.Sprintf("⏰ %s बबुआ, आज के 15-मिनट के अभ्यास बाकी बा! जल्दी से शुरू करा।", name)
		case "PUNJABI":
			return fmt.Sprintf("⏰ %s ਪੁੱਤਰ, ਅੱਜ ਦਾ 15-ਮਿੰਟ ਅਭਿਆਸ ਬਾਕੀ ਹੈ! ਜਲਦੀ ਸ਼ੁਰੂ ਕਰੋ।", name)
		default:
			return fmt.Sprintf("⏰ %s बेटा, आज का 15-मिनट अभ्यास अभी बाकी है! जल्दी शुरू करें।", name)
		}

	case missedDays == 2:
		// दिन 2: स्ट्रीक बचाने की प्रेरणा
		streakText := ""
		if currentStreak > 1 {
			streakText = fmt.Sprintf("आपकी %d-दिन की मेडल स्ट्रीक टूटने वाली है! ", currentStreak)
		}
		return fmt.Sprintf("🔥 %s, %sआज ही अभ्यास पूरा करके अपनी स्ट्रीक जारी रखें।", name, streakText)

	case missedDays == 3:
		// दिन 3: अभिभावक के लिए सीधा अलर्ट
		return fmt.Sprintf("👨‍👩‍👦 *अभिभावक सूचना:* %s ने पिछले 3 दिनों से 15-मिनट अभ्यास नहीं किया है। कृपया बच्चे को आज अभ्यास के लिए प्रेरित करें।", name)

	default:
		// 4 या अधिक दिन: री-एंगेजमेंट और सपोर्ट
		return fmt.Sprintf("📢 %s के अभिभावक जी, अभ्यास में निरंतरता ही सफलता की कुंजी है। क्या पढ़ाई में कोई समस्या आ रही है? सहायता के लिए 9664006651 पर संपर्क करें।", name)
	}
}
