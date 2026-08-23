package monitor

import (
	"fmt"
	"strings"
)

type MindReader struct{}

func NewMindReader() *MindReader {
	return &MindReader{}
}

func (m *MindReader) ProbeCheekyStudent(studentName, message string) string {
	lower := strings.ToLower(message)
	if strings.Contains(lower, "क्यों बताऊँ") || strings.Contains(lower, "तुम जानो") || strings.Contains(lower, "नहीं बताऊंगा") {
		return fmt.Sprintf("अरे %s! 🕵️‍♂️ जासूसी मूड छोड़िए और आज का 15-मिनट अभ्यास भेजिए ताकि आपका मेडल सुरक्षित रहे!", studentName)
	}
	return "बहुत बढ़िया! चलिए आज के 15-मिनट अभ्यास पर ध्यान देते हैं।"
}
