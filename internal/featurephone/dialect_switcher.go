package featurephone

import (
	"database/sql"
	"log"
)

type DialectSwitcher struct {
	db *sql.DB
}

func NewDialectSwitcher(db *sql.DB) *DialectSwitcher {
	return &DialectSwitcher{db: db}
}

func (d *DialectSwitcher) SwitchDialect(studentUID, keypadInput string) (string, string) {
	dialect := "HINDI"
	prompt := "राम राम बेटा! आज का अभ्यास शुरू करते हैं।"

	switch keypadInput {
	case "1":
		dialect = "HINDI"
		prompt = "राम राम बेटा! आज का अभ्यास शुरू करते हैं।"
	case "2":
		dialect = "RAJASTHANI_BAGRI"
		prompt = "राम राम बेटा! आज रो अभ्यास शुरू करां।"
	case "3":
		dialect = "HARYANVI"
		prompt = "राम राम भाई! आज का ताऊ कै हिसाब तै अभ्यास शुरू करां।"
	case "4":
		dialect = "BHOJPURI"
		prompt = "प्रणाम भैया! आज के पढ़ाई शुरू कइल जाव।"
	case "5":
		dialect = "PUNJABI"
		prompt = "ਸਤਿ ਸ੍ਰੀ ਅਕਾਲ ਬੇਟਾ! ਅੱਜ ਦਾ ਅਭਿਆਸ ਸ਼ੁਰੂ ਕਰਦੇ ਹਾਂ।"
	case "6":
		dialect = "BENGALI"
		prompt = "নমস্কার সোনা! আজকের অনুশীলন शुरू করা যাক।"
	case "7":
		dialect = "MARATHI"
		prompt = "नमस्कार बाळ! आजचा सराव सुरू करूया."
	case "8":
		dialect = "TELUGU"
		prompt = "నమస్కారం బాబు! ఈరోజు అభ్యాసం ప్రారంభిద్దాం."
	case "9":
		dialect = "TAMIL"
		prompt = "வணக்கம் தம்பி! இன்றைய பயிற்சியைத் தொடங்கலாம்."
	case "0":
		dialect = "ENGLISH"
		prompt = "Hello champion! Let's start today's practice session."
	default:
		dialect = d.getPreviousSessionDialect(studentUID)
		prompt = "राम राम बेटा! आज का अभ्यास शुरू करते हैं।"
	}

	go func() {
		_, _ = d.db.Exec(`UPDATE students SET last_used_dialect = $1, last_dialect_switched_at = NOW() WHERE uid = $2`, dialect, studentUID)
	}()

	log.Printf("🌐 [डायनामिक भाषा स्विच] छात्र %s ने सत्र भाषा '%s' चुनी।", studentUID, dialect)
	return dialect, prompt
}

func (d *DialectSwitcher) getPreviousSessionDialect(studentUID string) string {
	var dialect string
	query := `SELECT COALESCE(preferred_dialect, 'HINDI') FROM users u JOIN onboarding_sessions o ON u.phone = o.phone WHERE u.uid_badge = $1`
	err := d.db.QueryRow(query, studentUID).Scan(&dialect)
	if err != nil {
		return "HINDI"
	}
	return dialect
}
