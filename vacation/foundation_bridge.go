package vacation

import (
	"fmt"
	"strings"
)

type BridgeProgramInfo struct {
	StudentName     string   `json:"student_name"`
	CurrentGrade    int      `json:"current_grade"`
	TargetGrade     int      `json:"target_grade"`
	TransitionPhase string   `json:"transition_phase"`
	KeyFocusAreas   []string `json:"key_focus_areas"`
	PromptPayload   string   `json:"prompt_payload"`
}

type FoundationBridgeService struct{}

func NewFoundationBridgeService() *FoundationBridgeService {
	return &FoundationBridgeService{}
}

// BuildBridgePrompt छात्र की वर्तमान और अगली कक्षा के बीच 15-मिनट का फाउंडेशन ब्रिज प्रॉम्प्ट तैयार करता है
func (f *FoundationBridgeService) BuildBridgePrompt(studentName string, currentGrade, targetGrade int) *BridgeProgramInfo {
	cleanName := strings.TrimSpace(studentName)
	if cleanName == "" {
		cleanName = "विद्यार्थी"
	}

	// सुरक्षा: डिफ़ॉल्ट रूप से अगली कक्षा (Current + 1)
	if targetGrade <= currentGrade || targetGrade <= 0 {
		targetGrade = currentGrade + 1
	}
	if currentGrade < 1 {
		currentGrade = 1
	}
	if targetGrade > 12 {
		targetGrade = 12
	}

	var focusAreas []string
	var transitionPhase string

	switch {
	case currentGrade <= 5:
		transitionPhase = "प्राथमिक से उच्च-प्राथमिक संक्रमण (Primary to Middle Foundation)"
		focusAreas = []string{
			"मूलभूत गणित (भिन्न, दशमलव और व्यावहारिक गणना)",
			"भाषा अभिव्यक्ति और स्वतंत्र पठन-लेखन",
			"पर्यावरण और दैनिक विज्ञान के बुनियादी 'क्यों और कैसे'",
		}

	case currentGrade >= 6 && currentGrade <= 8:
		transitionPhase = "माध्यमिक स्तर की नींव (Middle to High School Bridge)"
		focusAreas = []string{
			"बीजगणित (Algebra) और ज्यामिति के मूल नियम",
			"भौतिकी और रसायन विज्ञान के आधारभूत सिद्धांत",
			"तार्किक समझ और चरणबद्ध समस्या निवारण",
		}

	case currentGrade == 9 || currentGrade == 10:
		transitionPhase = "बोर्ड एवं स्ट्रीम चयन ब्रिज (Secondary to Senior Secondary Bridge)"
		focusAreas = []string{
			"कक्षा 10 बोर्ड दक्षता और जटिल कॉन्सेप्ट्स का सरलीकरण",
			"कक्षा 11 के मुख्य विषयों (Science/Commerce/Arts) का बुनियादी पूर्वाभ्यास",
			"समय प्रबंधन और पिछले वर्षों के महत्वपूर्ण पैटर्न",
		}

	default: // Grade 11 to 12
		transitionPhase = "करियर व प्रतियोगी परीक्षा तैयारी (Senior Secondary Mastery)"
		focusAreas = []string{
			"कक्षा 11 के कमजोर अध्यायों का त्वरित रिवीज़न",
			"कक्षा 12 बोर्ड परीक्षा और प्रवेश परीक्षाओं (JEE/NEET/CUET) का अग्रिम तालमेल",
			"उन्नत विश्लेषणात्मक और व्यावहारिक अभ्यास",
		}
	}

	prompt := fmt.Sprintf(
		"=== फाउंडेशन ब्रिज प्रोग्राम (FOUNDATION BRIDGE: कक्षा %d ➔ कक्षा %d) ===\n"+
			"• विद्यार्थी: %s\n"+
			"• संक्रमण चरण: %s\n"+
			"• मुख्य फोकस क्षेत्र: %s\n\n"+
			"【AI शिक्षण निर्देश】\n"+
			"1. छात्र को कक्षा %d के भार से डराए बिना, केवल 15 मिनट में कक्षा %d के एक बुनियादी नियम का अग्रिम परिचय कराएं।\n"+
			"2. शिक्षण का माध्यम सरल, रोचक और व्यावहारिक उदाहरणों से युक्त रखें।\n"+
			"3. अंत में छात्र से हाथ से लिखकर हल करने वाला केवल 1 छोटा प्रश्न पूछें ताकि उसका आत्मविश्वास बढ़े।\n"+
			"4. टोन हमेशा एक मार्गदर्शक और प्रेरक स्थानीय गुरुजी जैसी रखें।\n"+
			"==================================================================",
		currentGrade, targetGrade,
		cleanName,
		transitionPhase,
		strings.Join(focusAreas, ", "),
		targetGrade, targetGrade,
	)

	return &BridgeProgramInfo{
		StudentName:     cleanName,
		CurrentGrade:    currentGrade,
		TargetGrade:     targetGrade,
		TransitionPhase: transitionPhase,
		KeyFocusAreas:   focusAreas,
		PromptPayload:   prompt,
	}
}
