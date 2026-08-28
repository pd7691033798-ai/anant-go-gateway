package gateway

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type ContactEntry struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type PersonalContactsManager struct {
	db *sql.DB
}

func NewPersonalContactsManager(db *sql.DB) *PersonalContactsManager {
	return &PersonalContactsManager{db: db}
}

// sanitizePhoneNumber 10 अंकों का शुद्ध मोबाइल नंबर निकालता है
func sanitizePhoneNumber(phone string) string {
	clean := strings.TrimPrefix(phone, "+91")
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, "(", "")
	clean = strings.ReplaceAll(clean, ")", "")
	if len(clean) >= 10 {
		return clean[len(clean)-10:]
	}
	return clean
}

func (p *PersonalContactsManager) SyncPhonebookContacts(w http.ResponseWriter, r *http.Request) {
	var list []ContactEntry
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		http.Error(w, `{"error": "अमान्य डेटा फॉर्मेट"}`, http.StatusBadRequest)
		return
	}

	tx, err := p.db.Begin()
	if err != nil {
		log.Printf("❌ [DB Error] ट्रांजेक्शन शुरू नहीं हो सका: %v", err)
		http.Error(w, `{"error": "सर्वर त्रुटि"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO personal_whitelist (phone, contact_name, added_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (phone) DO UPDATE SET contact_name = EXCLUDED.contact_name`)
	if err != nil {
		log.Printf("❌ [DB Error] प्रिपेयर्ड स्टेटमेंट फेल: %v", err)
		http.Error(w, `{"error": "क्वेरी त्रुटि"}`, http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	syncedCount := 0
	for _, item := range list {
		clean := sanitizePhoneNumber(item.Phone)
		if len(clean) == 10 {
			if _, err := stmt.Exec(clean, item.Name); err == nil {
				syncedCount++
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("❌ [DB Error] कमिट फेल हुआ: %v", err)
		http.Error(w, `{"error": "डेटा सेव नहीं हो सका"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("📱 [फ़ोनबुक सिंक] %d/%d पर्सनल कॉन्टैक्ट्स सुरक्षित सिंक किए गए।", syncedCount, len(list))
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "SUCCESS", "synced": true}`))
}

func (p *PersonalContactsManager) IsPersonalContact(phone string) (bool, string) {
	cleanPhone := sanitizePhoneNumber(phone)
	if len(cleanPhone) != 10 {
		return false, ""
	}

	var contactName string
	query := `SELECT contact_name FROM personal_whitelist WHERE phone = $1 LIMIT 1`
	err := p.db.QueryRow(query, cleanPhone).Scan(&contactName)
	if err == nil {
		return true, contactName
	}
	return false, ""
}
