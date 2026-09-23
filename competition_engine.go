package main

import (
	"errors"
	"fmt"
	"time"
)

// परीक्षा श्रेणियाँ (Competition Categories)
type ExamCategory string

const (
	ExamSainikSchool ExamCategory = "SAINIK_SCHOOL"
	ExamNavodaya     ExamCategory = "NAVODAYA"
	ExamJEEPre       ExamCategory = "JEE_PRE"      // पहले JEE प्री (Main)
	ExamJEEMains     ExamCategory = "JEE_MAINS"    // फिर JEE मेन्स / एडवांस
	ExamStateBoard   ExamCategory = "STATE_BOARD"  // 11वीं और 12वीं के सभी स्टेट बोर्ड्स
)

// डेटा के प्रकार (A to Z Study Material & Updates)
type ResourceType string

const (
	ResourcePreviousPaper ResourceType = "PREVIOUS_YEAR_PAPER"
	ResourceSyllabus      ResourceType = "SYLLABUS"
	ResourceMockTest      ResourceType = "MOCK_TEST"
	ResourceLiveUpdate    ResourceType = "LIVE_UPDATED_CONTENT" // हमेशा अपडेटेड और फ्रेश डेटा
)

// परीक्षा डेटा संरचना (Exam Content Structure)
type ExamContent struct {
	Category     ExamCategory
	TargetName   string       // जैसे: 'JEE Pre 2026' या 'UP Board Class 12 Science'
	ResourceType ResourceType
	Title        string
	ContentData  string       // प्रश्न, उत्तर, थ्योरी या अपडेटेड नोट्स
	Version      string       // संस्करण ताकि पुराना मटीरियल न रहे
	LastUpdated  time.Time    // यह सुनिश्चित करने के लिए कि डेटा हमेशा ताज़ा रहे
}

// कंपटीशन इंजन (Competition Engine) - नाम अपडेट किया गया
type CompetitionEngine struct {
	Database map[string][]ExamContent
}

// नया कंपटीशन इंजन शुरू करना
func NewCompetitionEngine() *CompetitionEngine {
	engine := &CompetitionEngine{
		Database: make(map[string][]ExamContent),
	}
	engine.LoadInitialData()
	return engine
}

// शुरुआती डेटा लोड करना (A to Z Content + Live Updates)
func (ce *CompetitionEngine) LoadInitialData() {
	now := time.Now()

	// 1. सैनिक स्कूल
	ce.Database[string(ExamSainikSchool)] = append(ce.Database[string(ExamSainikSchool)], ExamContent{
		Category:     ExamSainikSchool,
		TargetName:   "Class 6 & 9 Sainik Entrance",
		ResourceType: ResourcePreviousPaper,
		Title:        "Sainik School Previous Year Papers (Latest)",
		ContentData:  "गणित, सामान्य ज्ञान, बुद्धिमत्ता और भाषा के नवीनतम पैटर्न पर आधारित प्रश्न और हल।",
		Version:      "v2026.1",
		LastUpdated:  now,
	})

	// 2. नवोदय विद्यालय
	ce.Database[string(ExamNavodaya)] = append(ce.Database[string(ExamNavodaya)], ExamContent{
		Category:     ExamNavodaya,
		TargetName:   "Navodaya Class 6 Selection Test",
		ResourceType: ResourceMockTest,
		Title:        "Navodaya Mental Ability Mock Test",
		ContentData:  "मानसिक योग्यता परीक्षण (Mental Ability) के महत्वपूर्ण प्रश्न और शॉर्टकट ट्रिक्स।",
		Version:      "v2026.1",
		LastUpdated:  now,
	})

	// 3. JEE प्री और मेन्स
	ce.Database[string(ExamJEEPre)] = append(ce.Database[string(ExamJEEPre)], ExamContent{
		Category:     ExamJEEPre,
		TargetName:   "JEE Pre (Main) Foundation",
		ResourceType: ResourceSyllabus,
		Title:        "JEE Pre Physics & Math Fundamentals",
		ContentData:  "एनसीईआरटी आधारित बुनियादी अवधारणाएं और पिछले वर्षों के रुझान।",
		Version:      "v2026.1",
		LastUpdated:  now,
	})

	ce.Database[string(ExamJEEMains)] = append(ce.Database[string(ExamJEEMains)], ExamContent{
		Category:     ExamJEEMains,
		TargetName:   "JEE Mains Advanced Level",
		ResourceType: ResourceLiveUpdate,
		Title:        "JEE Mains High-Yield Formulas & Live Updates",
		ContentData:  "नवीनतम परीक्षा सत्र के अनुसार अपडेट किए गए प्रश्न और ट्रिक्स।",
		Version:      "v2026.1",
		LastUpdated:  now,
	})

	// 4. स्टेट बोर्ड्स (11वीं और 12वीं - सभी 28 राज्यों की सभी स्ट्रीम्स)
	ce.Database[string(ExamStateBoard)] = append(ce.Database[string(ExamStateBoard)], ExamContent{
		Category:     ExamStateBoard,
		TargetName:   "All States Class 11 & 12 (Science/Commerce/Arts)",
		ResourceType: ResourceLiveUpdate,
		Title:        "State Board Quick Revision Notes",
		ContentData:  "सभी 28 राज्यों के बोर्ड पैटर्न के अनुसार चैप्टर-वाइज महत्वपूर्ण प्रश्न।",
		Version:      "v2026.1",
		LastUpdated:  now,
	})
}

// छात्र द्वारा पूछे गए सवाल या खोज का जवाब देने का फंक्शन
func (ce *CompetitionEngine) QueryExamKnowledge(category ExamCategory) ([]ExamContent, error) {
	items, exists := ce.Database[string(category)]
	if !exists {
		return nil, errors.New("इस परीक्षा श्रेणी से संबंधित डेटा अभी उपलब्ध नहीं है")
	}
	return items, nil
}

// नया और अपडेटेड डेटा जोड़ने का फंक्शन (ताकि डेटा कभी पुराना न हो)
func (ce *CompetitionEngine) AddOrUpdateContent(category ExamCategory, target, title, data, version string, rType ResourceType) {
	newItem := ExamContent{
		Category:     category,
		TargetName:   target,
		ResourceType: rType,
		Title:        title,
		ContentData:  data,
		Version:      version,
		LastUpdated:  time.Now(), // हमेशा ताज़ा टाइमस्टैम्प ताकि कोई पुराना मटीरियल न रहे
	}
	catKey := string(category)
	ce.Database[catKey] = append(ce.Database[catKey], newItem)
	fmt.Printf("[Competition Engine Update]: '%s' (Version: %s) सफलतापूर्वक अपडेट और जोड़ा गया!\n", title, version)
}

func main() {
	// कंपटीशन इंजन शुरू करना
	competitionEngine := NewCompetitionEngine()

	// उदाहरण: एक नया अपडेटेड मॉक टेस्ट या स्टडी मटीरियल जोड़ना (कोचिंग से बेहतर और फ्रेश डेटा)
	competitionEngine.AddOrUpdateContent(
		ExamStateBoard, 
		"Class 12 Science (All State Boards)", 
		"Physics Chapter 1 to 5 Formula Sheet 2026", 
		"भौतिक विज्ञान के सभी नवीनतम सूत्र और व्युत्पत्ति।", 
		"v2026.2.0",
		ResourceLiveUpdate,
	)

	// छात्र द्वारा परीक्षा ज्ञान की खोज (Query)
	fmt.Println("\n--- छात्र खोज परिणाम (Competition Engine Query) ---")
	results, err := competitionEngine.QueryExamKnowledge(ExamStateBoard)
	if err != nil {
		fmt.Println("त्रुटि:", err)
		return
	}

	for _, res := range results {
		fmt.Printf("टारगेट: %s\nशीर्षक: %s\nवर्जन: %s\nकंटेंट: %s\nअंतिम अपडेट: %s\n--------------------\n", 
			res.TargetName, res.Title, res.Version, res.ContentData, res.LastUpdated.Format("2006-01-02 15:04:05"))
	}
}

