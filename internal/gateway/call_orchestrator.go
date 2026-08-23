package gateway

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"
)

const OfficialGatewayNumber = "9664006651"

type CallOrchestrator struct {
	db              *sql.DB
	personalManager *PersonalContactsManager
}

func NewCallOrchestrator(db *sql.DB, pm *PersonalContactsManager) *CallOrchestrator {
	return &CallOrchestrator{db: db, personalManager: pm}
}

func (co *CallOrchestrator) HandleMissCall(w http.ResponseWriter, r *http.Request) {
	incomingPhone := strings.TrimPrefix(r.URL.Query().Get("phone"), "+91")

	if isPersonal, name := co.personalManager.IsPersonalContact(incomingPhone); isPersonal {
		log.Printf("👤 [पर्सनल कॉल: %s (%s)] सामान्य घंटी बजेगी, AI दूर रहेगा।", name, incomingPhone)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"action": "ALLOW_NORMAL_RING", "trigger_ai": false}`))
		return
	}

	var studentName, dialect string
	var totalChildren int
	query := `
		SELECT u.name, COALESCE(u.total_children, 1), COALESCE(o.preferred_dialect, 'HINDI')
		FROM users u
		LEFT JOIN onboarding_sessions o ON u.phone = o.phone
		WHERE u.phone = $1 AND u.is_active = TRUE`

	err := co.db.QueryRow(query, incomingPhone).Scan(&studentName, &totalChildren, &dialect)
	if err != nil {
		log.Printf("🛡️ [ज़ीरो-रिस्क] %s अनजान नंबर है। शांत रहें।", incomingPhone)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"action": "IGNORE_SILENTLY", "trigger_ai": false}`))
		return
	}

	log.Printf("🎓 [सत्यापित छात्र] %s (%s) ➔ AI कॉल बैक", studentName, incomingPhone)
	go co.TriggerAICall(incomingPhone, studentName, totalChildren, dialect)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"action": "SILENT_DROP_AND_TRIGGER_AI", "trigger_ai": true}`))
}

func (co *CallOrchestrator) TriggerAICall(phone, name string, children int, dialect string) {
	time.Sleep(5 * time.Second)
	log.Printf("🤖 [AI मास्टरजी लाइव] %s पर कॉल कनेक्ट हुई | बच्चे: %d | भाषा: %s", phone, children, dialect)
}
