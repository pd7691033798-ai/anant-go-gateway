package vacation

import "fmt"

type PacingService struct{}

func NewPacingService() *PacingService {
	return &PacingService{}
}

func (p *PacingService) BuildPacedTaskPrompt(studentName, track string, currentDay, totalDays int) string {
	stage := "स्टेज 1: आधारभूत समझ"
	ratio := float64(currentDay) / float64(totalDays)

	if ratio > 0.75 {
		stage = "स्टेज 4: अंतिम मास्टरपीस पोर्टफोलियो"
	} else if ratio > 0.50 {
		stage = "स्टेज 3: एडवांस्ड प्रोजेक्ट निर्माण"
	} else if ratio > 0.25 {
		stage = "स्टेज 2: कोर थ्योरी व विस्तृत स्केच"
	}

	return fmt.Sprintf("विद्यार्थी %s | ट्रैक: %s | दिन %d/%d (%s) | कल के लिए 15-मिनट का अगला टास्क दें।",
		studentName, track, currentDay, totalDays, stage)
}
