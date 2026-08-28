package vacation

import (
	"fmt"
	"strings"
)

type PacingStage string

const (
	StageFoundation PacingStage = "स्टेज 1: आधारभूत समझ एवं जिज्ञासा (Foundation & Basics)"
	StageCoreTheory PacingStage = "स्टेज 2: कोर थ्योरी, आरेख एवं विस्तृत स्केच (Core Concept & Sketching)"
	StageAdvanced   PacingStage = "स्टेज 3: एडवांस्ड प्रोजेक्ट निर्माण (Hands-on Application)"
	StageMasterpiece PacingStage = "स्टेज 4: अंतिम मास्टरपीस एवं प्रस्तुति (Portfolio & Final Showcase)"
)

type PacedTaskInfo struct {
	StudentName     string      `json:"student_name"`
	Track           string      `json:"track"`
	CurrentDay      int         `json:"current_day"`
	TotalDays       int         `json:"total_days"`
	ProgressPercent float64     `json:"progress_percent"`
	CurrentStage    PacingStage `json:"current_stage"`
	StageDirective  string      `json:"stage_directive"`
	PromptPayload   string      `json:"prompt_payload"`
}

type PacingService struct{}

func NewPacingService() *PacingService {
	return &PacingService{}
}

// BuildPacedTaskPrompt छुट्टियों के दिनों और प्रगति अनुपात के अनुसार दैनिक गति और AI प्रॉम्प्ट निर्धारित करता है
func (p *PacingService) BuildPacedTaskPrompt(studentName, track string, currentDay, totalDays int) *PacedTaskInfo {
	cleanName := strings.TrimSpace(studentName)
	if cleanName == "" {
		cleanName = "विद्यार्थी"
	}

	cleanTrack := strings.TrimSpace(track)
	if cleanTrack == "" {
		cleanTrack = "रचनात्मक व्यावहारिक अभ्यास"
	}

	// सुरक्षा: शून्य या अमान्य दिनों से विभाजन रोकना
	if totalDays <= 0 {
		totalDays = 30
	}
	if currentDay < 1 {
		currentDay = 1
	}
	if currentDay > totalDays {
		currentDay = totalDays
	}

	ratio := float64(currentDay) / float64(totalDays)
	progressPercent := ratio * 100.0

	var stage PacingStage
	var directive string

	switch {
	case ratio > 0.75:
		stage = StageMasterpiece
		directive = "अवकाश का अंतिम चरण: बच्चे द्वारा अब तक सीखी गई बातों का एक सुंदर सारांश चार्ट या फाइनल पोर्टफोलियो पेज तैयार करवाएँ।"
	case ratio > 0.50:
		stage = StageAdvanced
		directive = "व्यावहारिक अनुप्रयोग चरण: विषय के सिद्धांतों को किसी वास्तविक मॉडल, प्रयोग या केस-स्टडी से जोड़कर हल करवाएँ।"
	case ratio > 0.25:
		stage = StageCoreTheory
		directive = "अवधारणा सुदृढ़ीकरण: मुख्य सूत्रों, आरेखों (Diagrams) और चरणबद्ध प्रक्रियाओं को हाथ से लिखवाएँ।"
	default:
		stage = StageFoundation
		directive = "बुनियादी परिचय: विषय के प्रति बच्चे की रुचि जगाने के लिए सरल, रोचक और बुनियादी तथ्य व प्रश्न दें।"
	}

	prompt := fmt.Sprintf(
		"=== गतिशील गति एवं चरणबद्ध अवकाश कार्य (DYNAMIC VACATION PACING) ===\n"+
			"• विद्यार्थी: %s\n"+
			"• लर्निंग ट्रैक: '%s'\n"+
			"• अवकाश प्रगति: दिन %d / %d (प्रगति: %.1f%%)\n"+
			"• वर्तमान चरण: %s\n\n"+
			"【AI शिक्षण निर्देश】\n"+
			"1. %s\n"+
			"2. आज के लिए केवल 15 मिनट का एकल, केंद्रित और हाथ से लिखने योग्य व्यावहारिक अभ्यास तैयार करें।\n"+
			"3. कार्य ऐसा हो जो बच्चे पर मानसिक बोझ न डाले और छुट्टियों के आनंद के साथ ज्ञान बढ़ाए।\n"+
			"==================================================================",
		cleanName, cleanTrack, currentDay, totalDays, progressPercent, stage, directive,
	)

	return &PacedTaskInfo{
		StudentName:     cleanName,
		Track:           cleanTrack,
		CurrentDay:      currentDay,
		TotalDays:       totalDays,
		ProgressPercent: progressPercent,
		CurrentStage:    stage,
		StageDirective:  directive,
		PromptPayload:   prompt,
	}
}

