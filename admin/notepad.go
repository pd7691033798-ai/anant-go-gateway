package admin

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type NotepadEngine struct {
	db *sql.DB
}

func NewNotepadEngine(db *sql.DB) *NotepadEngine {
	return &NotepadEngine{db: db}
}

// ProcessNotepadInput - यह नोटपैड से लिखे गए हर विचार को सुरक्षित तरीके से कैप्चर करता है
func (ne *NotepadEngine) ProcessNotepadInput(rawThought string) (string, error) {
	thought := strings.TrimSpace(rawThought)
	if thought == "" {
		return "", fmt.Errorf("[SYNTAX ERROR] नोटपैड खाली है: कृपया स्पष्ट निर्देश लिखें।")
	}

	// 1. सुरक्षा जाँच (SQL Injection और खतरनाक कमांड्स को ब्लॉक करना)
	thoughtUpper := strings.ToUpper(thought)
	if strings.Contains(thoughtUpper, "DROP TABLE") || 
	   strings.Contains(thoughtUpper, "DELETE FROM") || 
	   strings.Contains(thoughtUpper, "TRUNCATE") ||
	   strings.Contains(thoughtUpper, "EXEC") {
		log.Printf("[SECURITY ALERT] Dangerous command blocked from notepad: %s", thought)
		return "", fmt.Errorf("[SECURITY VIOLATION] खतरनाक या अनधिकृत डेटाबेस कमांड ब्लॉक कर दी गई है।")
	}

	thoughtLower := strings.ToLower(thought)

	// 2. कैटेगरी पहचानना और सुरक्षित रूप से डेटाबेस में भेजना
	if strings.Contains(thoughtLower, "feature") || strings.Contains(thoughtLower, "जोड़ो") || strings.Contains(thoughtLower, "add") {
		return ne.saveRuleToDB(thought, "FEATURE")
	}

	if strings.Contains(thoughtLower, "price") || strings.Contains(thoughtLower, "discount") || strings.Contains(thoughtLower, "फीस") || strings.Contains(thoughtLower, "डिस्काउंट") {
		return ne.saveRuleToDB(thought, "PRICING")
	}

	// 3. अगर सिस्टम को निर्देश समझ न आए (तेरी माँगी हुई एरर)
	log.Printf("[SYNTAX UNRECOGNIZED] Instruction not matched: %s", thought)
	return "", fmt.Errorf("[FILE NOT FOUND / SYNTAX UNRECOGNIZED] सिस्टम इस निर्देश को किसी भी इंजन या फोल्डर में मैप नहीं कर सका। कृपया वाक्य या नियम को सुधार कर दोबारा लिखें।")
}

// सुरक्षित रूप से डेटाबेस में नियम सेव करने का फंक्शन (Parameterized Query)
func (ne *NotepadEngine) saveRuleToDB(ruleText, ruleType string) (string, error) {
	query := `INSERT INTO dynamic_rules (rule_text, rule_type, is_active, created_at) VALUES ($1, $2, true, NOW())`
	
	// यहाँ सीधे स्ट्रिंग जोड़ने के बजाय $1 का उपयोग किया गया है ताकि कोई SQL Injection न हो सके
	_, err := ne.db.Exec(query, ruleText, ruleType)
	if err != nil {
		log.Printf("[DATABASE ERROR] Failed to save rule: %v", err)
		return "", fmt.Errorf("डेटाबेस में सेव करते वक्त त्रुटि आई: %v", err)
	}

	return fmt.Sprintf("✅ [SUCCESS] %s नियम को नोटपैड से सफलताપૂર્વक सुरक्षित रूप से दर्ज कर लिया गया है!", ruleType), nil
}
