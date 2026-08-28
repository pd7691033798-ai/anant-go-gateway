package student

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"
)

type SkillDomain string

const (
	DomainNextGradeAcademic SkillDomain = "ACADEMIC_NEXT_GRADE"
	DomainComputerCoding    SkillDomain = "COMPUTER_AND_CODING"
	DomainMechanicalTech    SkillDomain = "MECHANICAL_AND_PRACTICAL"
	DomainCommerceFinance   SkillDomain = "COMMERCE_AND_FINANCE"
)

// AutoLearningBridge इंटरफ़ेस - यह आगे आने वाले Auto-Learning Engine से सीधे कनेक्ट होगा
type AutoLearningBridge interface {
	RecordStudentInitialInterest(ctx context.Context, studentUID string, domain SkillDomain) error
	GetAdaptiveDomainRecommendation(ctx context.Context, studentUID string) (SkillDomain, bool)
}

type SkillDiscoveryEngine struct {
	db                 *sql.DB
	autoLearningEngine AutoLearningBridge // ऑटो-लर्निंग का प्लग-इन हुक
}

// NewSkillDiscoveryEngine नया इंस्टेंस बनाता है (autoLearningEngine वैकल्पिक / optional है)
func NewSkillDiscoveryEngine(db *sql.DB, autoLearning AutoLearningBridge) *SkillDiscoveryEngine {
	return &SkillDiscoveryEngine{
		db:                 db,
		autoLearningEngine: autoLearning,
	}
}

// SetAutoLearningEngine बाद में ऑटो-लर्निंग इंजन को अटैच करने के लिए
func (s *SkillDiscoveryEngine) SetAutoLearningEngine(autoLearning AutoLearningBridge) {
	s.autoLearningEngine = autoLearning
}

func (s *SkillDiscoveryEngine) HandleSkillSelection(studentUID string, optionKey string) (string, SkillDomain) {
	studentUID = strings.TrimSpace(studentUID)
	optionKey = strings.TrimSpace(optionKey)

	var selectedDomain SkillDomain
	var audioResponse string

	switch optionKey {
	case "2":
		selectedDomain = DomainComputerCoding
		audioResponse = "शाबाश बेटा! कंप्यूटर कोडिंग से आप ऐप्स, गेम्स और वेबसाइट्स बना सकते हैं।"
	case "3":
		selectedDomain = DomainMechanicalTech
		audioResponse = "बहुत बढ़िया बेटा! मैकेनिकल में आप मशीनों और रोबोटिक्स के बारे में सीखते हैं।"
	case "4":
		selectedDomain = DomainCommerceFinance
		audioResponse = "शाबाश! कॉमर्स में आप बिज़नेस, CA और बैंकिंग के नियम सीखते हैं।"
	default:
		selectedDomain = DomainNextGradeAcademic
		audioResponse = "चलिए बेटा, आज हम आपकी अगली स्कूली कक्षा के मुख्य सूत्रों का ओरिएंटेशन शुरू करते हैं।"
	}

	// बैकग्राउंड में DB अपडेट + Auto-Learning Engine को डेटा पास करना
	if studentUID != "" {
		go func(uid string, domain SkillDomain) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 1. डेटाबेस में सुरक्षित अपडेट
			query := `
				UPDATE students 
				SET career_interest_domain = $1, 
				    last_skill_session_at = NOW(),
				    updated_at = NOW() 
				WHERE uid = $2`

			result, err := s.db.ExecContext(ctx, query, string(domain), uid)
			if err != nil {
				log.Printf("❌ [SkillDiscoveryEngine] DB अपडेट त्रुटि (UID: %s, Domain: %s): %v", uid, domain, err)
			} else if rows, _ := result.RowsAffected(); rows == 0 {
				log.Printf("⚠️ [SkillDiscoveryEngine] छात्र UID नहीं मिला: %s", uid)
			}

			// 2. ऑटो-लर्निंग इंजन को सिग्नल भेजना (यदि इंजन अटैच्ड है)
			if s.autoLearningEngine != nil {
				if err := s.autoLearningEngine.RecordStudentInitialInterest(ctx, uid, domain); err != nil {
					log.Printf("⚠️ [SkillDiscoveryEngine] Auto-Learning सिंक एरर (UID: %s): %v", uid, err)
				} else {
					log.Printf("🧠 [Auto-Learning Link] छात्र %s का डोमेन '%s' ऑटो-लर्निंग मॉडल में फीड हो गया।", uid, domain)
				}
			}
		}(studentUID, selectedDomain)
	}

	return audioResponse, selectedDomain
}
