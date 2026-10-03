package admin

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type NotepadEngine struct {
	db        *sql.DB
	refiner   *RefinementEngine
}

func NewNotepadEngine(db *sql.DB) *NotepadEngine {
	return &NotepadEngine{
		db:      db,
		refiner: NewRefinementEngine(),
	}
}

// 1. यह मुख्य फंक्शन है जो तेरे नोटपैड के लिखे हुए हर शब्द को कैच करेगा
func (ne *NotepadEngine) ProcessNotepadInput(rawThought string) (string, error) {
	thought := strings.TrimSpace(rawThought)
	if thought == "" {
		return "", fmt.Errorf("[SYNTAX ERROR] फाइल खाली है: कृपया नोटपैड में निर्देश स्पष्ट लिखें।")
	}

	log.Printf("[NOTEPAD INPUT RECEIVED] Processing admin instruction...", )

	// 2. सुरक्षा और शुद्धता के लिए पहले इसे 'रिफाइनमेंट इंजन' से गुजारेंगे (Quality Gate)
	// (अगर तूने नोटपैड में कोई कोड या लॉजिक लिखा है, तो यह उसे स्कैन करेगा)
	_, refinementErr := ne.refiner.RefineSystemComponent("NotepadInstruction", thought)
	if refinementErr != nil {
		return "", fmt.Errorf("[REFINEMENT FAILED] आपके निर्देश में सिंटेक्स या लॉजिकल त्रुटि है: %v", refinementErr)
	}

	thoughtLower := strings.ToLower(thought)

	// 3. यह तय करना कि यह निर्देश किस कैटेगोरी में जाएगा
	if strings.Contains(thoughtLower, "feature") || strings.Contains(thoughtLower, "जोड़ो") || strings.Contains(thoughtLower, "add") {
		return ne.executeFeatureInjection(thought)
	}

	if strings.Contains(thoughtLower, "price") || strings.Contains(thoughtLower, "discount") || strings.Contains(thoughtLower, "फीस") || strings.Contains(thoughtLower, "डिस्काउंट") {
		return ne.executePricingInjection(thought)
	}

	// 4. अगर निर्देश समझ नहीं आया, तो जैसा तूने कहा था—फाइल नॉट फाउंड का एरर फेंको!
	log.Printf("[ERROR] Unrecognized instruction pattern: %s", thought)
	return "", fmt.Errorf("[FILE NOT FOUND / SYNTAX ERROR] सिस्टम इस निर्देश को किसी भी इंजन या फोल्डर में मैप नहीं कर सका। कृपया वाक्य या नियम को सुधार कर दोबारा लिखें।")
}

// नियम 1: फीचर को ऑटोमैटिक डायनेमिक डायरेक्ट्री में डालना
func (ne *NotepadEngine) executeFeatureInjection(instruction string) (string, error) {
	query := `INSERT INTO dynamic_rules (rule_text, rule_type, is_active, created_at) VALUES ($1, 'FEATURE', true, NOW())`
	_, err := ne.db.Exec(query, instruction)
	if err != nil {
		return "", fmt.Errorf("डेटाबेस में फीचर इंजेक्ट करते वक्त एरर आया: %v", err)
	}
	return "✅ [SUCCESS] नोटपैड के निर्देश को पढ़ लिया गया है। 'सिस्टम रिफाइनमेंट' पास हो चुका है और नया फीचर नियम डेटाबेस में लाइव हो गया है!", nil
}

// नियम 2: प्राइसिंग या डिस्काउंट नियम को ऑटोमैटिक सेट करना
func (ne *NotepadEngine) executePricingInjection(instruction string) (string, error) {
	query := `INSERT INTO dynamic_rules (rule_text, rule_type, is_active, created_at) VALUES ($1, 'PRICING', true, NOW())`
	_, err := ne.db.Exec(query, instruction)
	if err != nil {
		return "", fmt.Errorf("प्राइसिंग रूल सेव नहीं हो सका: %v", err)
	}
	return "💰 [SUCCESS] नया प्राइसिंग/डिस्काउंट नियम नोटपैड से सफलताપૂર્વक कैप्चर कर लिया गया है। ऑटो-पे इंजन अब इसे लागू करेगा!", nil
}
