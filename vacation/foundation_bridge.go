package vacation

import "fmt"

type FoundationBridgeService struct{}

func NewFoundationBridgeService() *FoundationBridgeService {
	return &FoundationBridgeService{}
}

func (f *FoundationBridgeService) BuildBridgePrompt(studentName string, currentGrade, targetGrade int) string {
	return fmt.Sprintf(
		"=== फाउंडेशन ब्रिज प्रोग्राम (कक्षा %d ➔ %d) ===\n"+
		"विद्यार्थी: %s\n"+
		"निर्देश: अगली कक्षा के कोर सिद्धांतों का 15-मिनट अग्रिम परिचय कराएं।",
		currentGrade, targetGrade, studentName,
	)
}
