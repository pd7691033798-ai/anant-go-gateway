package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

type StudentKundali struct {
	UIDBadge          string   `json:"uid_badge"`
	StudentName       string   `json:"student_name"`
	ClassLevel        int      `json:"class_level"`
	SchoolName        string   `json:"school_name"`
	ParentName        string   `json:"parent_name"`
	ParentPhone       string   `json:"parent_phone"`
	PlanType          string   `json:"plan_type"`
	NextPayableAmount float64  `json:"next_payable_amount"`
	PreferredDialect  string   `json:"preferred_dialect"`
	AccuracyPercent   float64  `json:"accuracy_percent"`
	WeakTopics        []string `json:"weak_topics"`
}

type KundaliSearchEngine struct {
	db *sql.DB
}

func NewKundaliSearchEngine(db *sql.DB) *KundaliSearchEngine {
	return &KundaliSearchEngine{db: db}
}

func (k *KundaliSearchEngine) HandleKundaliSearch(w http.ResponseWriter, r *http.Request) {
	searchTerm := strings.TrimSpace(r.URL.Query().Get("q"))
	searchTerm = strings.TrimPrefix(searchTerm, "+91")

	if searchTerm == "" {
		http.Error(w, `{"error": "बैच कोड या फोन नंबर दर्ज करें"}`, http.StatusBadRequest)
		return
	}

	query := `
		SELECT 
			COALESCE(u.uid_badge, ''),
			COALESCE(o.child_name, 'अज्ञात'),
			COALESCE(CAST(o.class_level AS INTEGER), 1),
			COALESCE(o.school_name, 'विद्यालय'),
			COALESCE(u.name, 'अभिभावक'),
			u.phone,
			COALESCE(o.preferred_dialect, 'HINDI')
		FROM users u
		LEFT JOIN onboarding_sessions o ON u.phone = o.phone
		WHERE u.uid_badge = $1 OR u.phone = $1
		LIMIT 1`

	var kData StudentKundali
	err := k.db.QueryRow(query, searchTerm).Scan(
		&kData.UIDBadge,
		&kData.StudentName,
		&kData.ClassLevel,
		&kData.SchoolName,
		&kData.ParentName,
		&kData.ParentPhone,
		&kData.PreferredDialect,
	)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "रिकॉर्ड नहीं मिला"}`))
		return
	}

	kData.PlanType = "BASIC"
	kData.NextPayableAmount = 299.25
	kData.AccuracyPercent = 85.0
	kData.WeakTopics = []string{"गणित", "विज्ञान"}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(kData)
}
