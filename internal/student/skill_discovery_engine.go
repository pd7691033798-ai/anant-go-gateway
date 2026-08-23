package student

import (
	"database/sql"
)

type SkillDomain string

const (
	DomainNextGradeAcademic SkillDomain = "ACADEMIC_NEXT_GRADE"
	DomainComputerCoding    SkillDomain = "COMPUTER_AND_CODING"
	DomainMechanicalTech    SkillDomain = "MECHANICAL_AND_PRACTICAL"
	DomainCommerceFinance   SkillDomain = "COMMERCE_AND_FINANCE"
)

type SkillDiscoveryEngine struct {
	db *sql.DB
}

func NewSkillDiscoveryEngine(db *sql.DB) *SkillDiscoveryEngine {
	return &SkillDiscoveryEngine{db: db}
}

func (s *SkillDiscoveryEngine) HandleSkillSelection(studentUID string, optionKey string) (string, SkillDomain) {
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

	go func() {
		_, _ = s.db.Exec(`UPDATE students SET career_interest_domain = $1, last_skill_session_at = NOW() WHERE uid = $2`, string(selectedDomain), studentUID)
	}()

	return audioResponse, selectedDomain
}
