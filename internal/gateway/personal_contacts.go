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

func (p *PersonalContactsManager) SyncPhonebookContacts(w http.ResponseWriter, r *http.Request) {
	var list []ContactEntry
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		http.Error(w, `{"error": "अमान्य डेटा"}`, http.StatusBadRequest)
		return
	}

	tx, _ := p.db.Begin()
	stmt, _ := tx.Prepare(`
		INSERT INTO personal_whitelist (phone, contact_name, added_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (phone) DO UPDATE SET contact_name = $2`)

	for _, item := range list {
		clean := strings.TrimPrefix(item.Phone, "+91")
		clean = strings.ReplaceAll(clean, " ", "")
		clean = strings.ReplaceAll(clean, "-", "")
		if len(clean) >= 10 {
			clean = clean[len(clean)-10:]
			_, _ = stmt.Exec(clean, item.Name)
		}
	}
	_ = stmt.Close()
	_ = tx.Commit()

	log.Printf("📱 [फ़ोनबुक सिंक] %d पर्सनल कॉन्टैक्ट्स सुरक्षित सिंक किए गए।", len(list))
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "SUCCESS"}`))
}

func (p *PersonalContactsManager) IsPersonalContact(phone string) (bool, string) {
	cleanPhone := strings.TrimPrefix(phone, "+91")
	if len(cleanPhone) >= 10 {
		cleanPhone = cleanPhone[len(cleanPhone)-10:]
	}
	var contactName string
	query := `SELECT contact_name FROM personal_whitelist WHERE phone = $1 LIMIT 1`
	err := p.db.QueryRow(query, cleanPhone).Scan(&contactName)
	if err == nil {
		return true, contactName
	}
	return false, ""
}
