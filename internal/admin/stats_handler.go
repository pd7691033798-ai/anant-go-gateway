package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type AdminStats struct {
	TotalActive        int `json:"total_active"`
	NewJoined          int `json:"new_joined"`
	ReferralsCompleted int `json:"referrals_completed"`
	NewReferrals       int `json:"new_referrals"`
}

type AdminStatsHandler struct {
	db *sql.DB
}

func NewAdminStatsHandler(db *sql.DB) *AdminStatsHandler {
	return &AdminStatsHandler{db: db}
}

func (h *AdminStatsHandler) GetStatsJSON(w http.ResponseWriter, r *http.Request) {
	var stats AdminStats

	// 1. कुल सक्रिय छात्र
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_active = TRUE`).Scan(&stats.TotalActive)

	// 2. नए जुड़े बच्चे (पिछले 7 दिनों में)
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '7 days'`).Scan(&stats.NewJoined)

	// 3. पूरे हुए रेफरल
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM referrals WHERE status = 'COMPLETED'`).Scan(&stats.ReferralsCompleted)

	// 4. नए पेंडिंग रेफरल
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM referrals WHERE status = 'PENDING'`).Scan(&stats.NewReferrals)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(stats)
}
