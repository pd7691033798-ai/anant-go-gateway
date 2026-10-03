package holiday

import (
	"database/sql"
	"fmt"
	"strings"
)

type SmartExamBank struct {
	db *sql.DB
}

func NewSmartExamBank(db *sql.DB) *SmartExamBank {
	return &SmartExamBank{db: db}
}

type ExamContentResult struct {
	Subject       string
	ExtractedTopics string
	IsFromCache   bool
	StudyPrompt   string
}

// ProcessExamInputAndCache: छात्र द्वारा भेजी गई फोटो/टेक्स्ट को प्रोसेस करता है और डेटाबेस में शेयरड नॉलेज बैंक बनाता है
func (sb *SmartExamBank) ProcessExamInputAndCache(phone string, subject string, rawTextOrImageCaption string) ExamContentResult {
	cleanSubject := strings.TrimSpace(subject)
	if cleanSubject == "" {
		cleanSubject = "सामान्य विज्ञान / गणित"
	}

	// 1. चेक करो कि क्या हमारे डेटाबेस में इस सब्जेक्ट/सिलेबस से मिलता-जुलता एग्जाम डेटा पहले से मौजूद है?
	var cachedTopics string
	queryCheck := `SELECT topics FROM shared_exam_knowledge_base WHERE subject ILIKE $1 LIMIT 1`
	
	if sb.db != nil {
		err := sb.db.QueryRow(queryCheck, "%"+cleanSubject+"%").Scan(&cachedTopics)
		if err == nil && cachedTopics != "" {
			// [कैश हिट] भविष्य का फायदा: डेटाबेस में पहले से बना-बनाया मटीरियल मिल गया!
			return ExamContentResult{
				Subject:       cleanSubject,
				ExtractedTopics: cachedTopics,
				IsFromCache:   true,
				StudyPrompt:   fmt.Sprintf("💡 (डेटाबेस से साझा मटीरियल)\nविषय: %s\nमुख्य टॉपिक्स: %s\n👉 इस पर आधारित आज का पहला स्मार्ट सवाल यह रहा:", cleanSubject, cachedTopics),
			}
		}
	}

	// 2. यदि नया मटीरियल है, तो AI/सिस्टम इसके टॉपिक्स खुद बनाएगा
	// (यहाँ फोटो OCR या टेक्स्ट से निकाले गए मुख्य टॉपिक्स हैं)
	newExtractedTopics := extractCoreTopics(rawTextOrImageCaption)

	// 3. इसे भविष्य के लिए डेटाबेस में सेव (सहेज) कर लो ताकि दूसरे बच्चों के काम आ सके
	if sb.db != nil {
		_, _ = sb.db.Exec(`
			INSERT INTO shared_exam_knowledge_base (subject, topics, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT DO NOTHING`, cleanSubject, newExtractedTopics)
	}

	return ExamContentResult{
		Subject:       cleanSubject,
		ExtractedTopics: newExtractedTopics,
		IsFromCache:   false,
		StudyPrompt:   fmt.Sprintf("✨ (नया सिलेबस विश्लेषित)\nविषय: %s\nमुख्य टॉपिक्स: %s\n👉 बिना रट्टा मारे, आइए इस टॉपिक को समझें:", cleanSubject, newExtractedTopics),
	}
}

// extractCoreTopics: भारी-भरकम थ्योरी से केवल काम के 2-3 मुख्य टॉपिक्स अलग करने का लॉजिक
func extractCoreTopics(rawText string) string {
	text := strings.ToLower(rawText)
	if strings.Contains(text, "math") || strings.Contains(text, "गणित") {
		return "द्विघात समीकरण (Quadratic Equations), त्रिकोणमिति सर्वसमिकाएँ (Trigonometric Identities)"
	} else if strings.Contains(text, "science") || strings.Contains(text, "विज्ञान") {
		return "प्रकाश का परावर्तन (Reflection of Light), रासायनिक अभिक्रियाएँ (Chemical Reactions)"
	}
	
	// डिफ़ॉल्ट टॉपिक यदि कुछ विशिष्ट न मिले
	return "कोर कॉन्सेप्ट, फॉर्मूला विश्लेषण, और स्टेप-मार्किंग अभ्यास"
}
