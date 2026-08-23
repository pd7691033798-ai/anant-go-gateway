package monitor

import "fmt"

type InactivityNudgeService struct{}

func NewInactivityNudgeService() *InactivityNudgeService {
	return &InactivityNudgeService{}
}

func (i *InactivityNudgeService) GenerateNudge(studentName string, missedDays int) string {
	switch missedDays {
	case 1:
		return fmt.Sprintf("⏰ शाम 6:00 बजे: %s बेटा, आज का 15-मिनट अभ्यास अभी बाकी है!", studentName)
	case 2:
		return fmt.Sprintf("🔥 रात 8:30 बजे: %s, आपकी 5-दिन की मेडल स्ट्रीक टूटने वाली है! जल्दी पेज भेजें।", studentName)
	default:
		return fmt.Sprintf("👨‍👩‍👦 अभिभावक सूचना: %s ने पिछले 3 दिनों से अभ्यास नहीं किया है। कृपया मार्गदर्शन करें।", studentName)
	}
}
