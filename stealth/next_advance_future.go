package stealth

import (
	"fmt"
	"strings"
)

type AcademicPhase string

const (
	PhaseEarlyExplorer AcademicPhase = "कक्षा 1-5: बुनियादी जिज्ञासा एवं रचनात्मक विकास (Early Exploration)"
	PhaseFoundation    AcademicPhase = "कक्षा 6-8: वैज्ञानिक दृष्टिकोण एवं समस्या समाधान (Foundation & Logic)"
	PhaseStreamChoice  AcademicPhase = "कक्षा 9-10: विषय चयन एवं बोर्ड दक्षता (Stream Selection & Aptitude)"
	PhaseCareerTarget  AcademicPhase = "कक्षा 11-12: करियर लक्ष्य एवं प्रतियोगी परीक्षा (Target Career & Entrances)"
)

type FutureMilestone struct {
	Timeline   string `json:"timeline"`
	FocusArea  string `json:"focus_area"`
	ActionItem string `json:"action_item"`
}

type FuturePathway struct {
	ChildName          string            `json:"child_name"`
	Grade              int               `json:"grade"`
	Hobby              string            `json:"hobby"`
	Phase              AcademicPhase     `json:"phase"`
	CoreCompetency     string            `json:"core_competency"`
	TargetMilestones   []FutureMilestone `json:"target_milestones"`
	RecommendedProject string            `json:"recommended_project"`
	WhatsAppReport     string            `json:"whatsapp_report"`
}

type NextAdvanceFutureEngine struct{}

func NewNextAdvanceFutureEngine() *NextAdvanceFutureEngine {
	return &NextAdvanceFutureEngine{}
}

// GenerateFuturePathway छात्र की कक्षा और हॉबी के आधार पर भविष्य का व्यापक रोडमैप तैयार करता है
func (n *NextAdvanceFutureEngine) GenerateFuturePathway(childName string, grade int, hobby string) *FuturePathway {
	cleanName := strings.TrimSpace(childName)
	if cleanName == "" {
		cleanName = "विद्यार्थी"
	}

	cleanHobby := strings.TrimSpace(hobby)
	if cleanHobby == "" {
		cleanHobby = "रोबोटिक्स और तार्किक चिंतन"
	}

	var phase AcademicPhase
	var coreCompetency string
	var milestones []FutureMilestone
	var project string

	switch {
	case grade <= 5:
		phase = PhaseEarlyExplorer
		coreCompetency = "पठन गति, संख्यात्मक गणना और व्यावहारिक रचनात्मकता"
		milestones = []FutureMilestone{
			{Timeline: "अगले 3 महीने", FocusArea: "मूलभूत गणित व भाषा", ActionItem: "दैनिक 15 मिनट बोलकर पढ़ने और गणना का अभ्यास।"},
			{Timeline: "अगले 6 महीने", FocusArea: "जिज्ञासा और हॉबी", ActionItem: fmt.Sprintf("%s से संबंधित छोटे व्यावहारिक मॉडल बनाना।", cleanHobby)},
		}
		project = fmt.Sprintf("घर पर आसान DIY मॉडल और %s पर आधारित चित्रों का संग्रह।", cleanHobby)

	case grade >= 6 && grade <= 8:
		phase = PhaseFoundation
		coreCompetency = "वैज्ञानिक दृष्टिकोण, गणितीय तर्क और समस्या समाधान"
		milestones = []FutureMilestone{
			{Timeline: "कक्षा 7-8", FocusArea: "ओलंपियाड व बेसिक साइंस", ActionItem: "गणित और विज्ञान में 'क्यों और कैसे' आधारित प्रश्न हल करना।"},
			{Timeline: "कक्षा 8 अंत", FocusArea: "टेक्निकल स्किल", ActionItem: fmt.Sprintf("%s के बुनियादी सिद्धांतों को समझना।", cleanHobby)},
		}
		project = fmt.Sprintf("%s पर आधारित एक कामकाजी साइंस प्रोजेक्ट या बेसिक लॉजिक बिल्डिंग।", cleanHobby)

	case grade >= 9 && grade <= 10:
		phase = PhaseStreamChoice
		coreCompetency = "बोर्ड परीक्षा में 90%+ स्कोर और सही स्ट्रीम (Science/Commerce/Arts) का चयन"
		milestones = []FutureMilestone{
			{Timeline: "कक्षा 9", FocusArea: "कॉन्सेप्ट स्पष्टता", ActionItem: "कठिन विषयों की दैनिक पुनरावृत्ति और अभ्यास।"},
			{Timeline: "कक्षा 10", FocusArea: "एप्टीट्यूड और स्ट्रीम", ActionItem: fmt.Sprintf("बोर्ड अंकों के साथ %s को करियर में बदलने के विकल्पों की पहचान।", cleanHobby)},
		}
		project = fmt.Sprintf("स्ट्रीम चयन के लिए एप्टीट्यूड टेस्ट और %s से जुड़े आधुनिक करियर की रिसर्च।", cleanHobby)

	default: // Grade 11-12
		phase = PhaseCareerTarget
		coreCompetency = "प्रतियोगी परीक्षा (JEE/NEET/CUET/NDA) और उच्च शिक्षा की तैयारी"
		milestones = []FutureMilestone{
			{Timeline: "कक्षा 11", FocusArea: "गहन अध्ययन", ActionItem: "समय प्रबंधन और पिछले वर्षों के प्रश्नपत्रों का विश्लेषण।"},
			{Timeline: "कक्षा 12", FocusArea: "लक्ष्य प्राप्ति", ActionItem: fmt.Sprintf("प्रवेश परीक्षा में शीर्ष रैंक हासिल करना और %s क्षेत्र में विशेषज्ञता।", cleanHobby)},
		}
		project = fmt.Sprintf("%s और अपने मुख्य विषय पर आधारित उन्नत पोर्टफोलियो व मॉक टेस्ट सीरीज।", cleanHobby)
	}

	// सुंदर व्हाट्सएप संदेश प्रारूप
	report := fmt.Sprintf(
		"🚀 *भविष्य का रोडमैप (Future Pathway)*\n\n"+
			"👤 *विद्यार्थी:* %s (कक्षा %d)\n"+
			"🎯 *वर्तमान चरण:* %s\n"+
			"💡 *हॉबी ट्रैक:* %s\n"+
			"⭐ *प्रमुख क्षमता विकास:* %s\n\n"+
			"📌 *आगामी मील के पत्थर:*\n"+
			"1. %s: %s\n"+
			"2. %s: %s\n\n"+
			"🛠️ *प्रस्तावित प्रोजेक्ट:* %s\n\n"+
			"✅ _अनंत अभ्यास: हर दिन 15 मिनट की मेहनत, सुरक्षित भविष्य की नींव!_",
		cleanName, grade, phase, cleanHobby, coreCompetency,
		milestones[0].Timeline, milestones[0].ActionItem,
		milestones[1].Timeline, milestones[1].ActionItem,
		project,
	)

	return &FuturePathway{
		ChildName:          cleanName,
		Grade:              grade,
		Hobby:              cleanHobby,
		Phase:              phase,
		CoreCompetency:     coreCompetency,
		TargetMilestones:   milestones,
		RecommendedProject: project,
		WhatsAppReport:     report,
	}
}
