package pricing

import (
	"database/sql"
	"time"
)

type DemoService struct {
	db *sql.DB
}

func NewDemoService(db *sql.DB) *DemoService {
	return &DemoService{db: db}
}

func (d *DemoService) Activate7DayDemo(phone, name string, grade int, state, district, dialect string) error {
	query := `INSERT INTO users (phone, name, grade, state, district, preferred_dialect, plan_tier, plan_expires_at)
	          VALUES ($1, $2, $3, $4, $5, $6, 'DEMO', NOW() + INTERVAL '7 days')
	          ON CONFLICT (phone) DO UPDATE 
	          SET plan_tier = 'DEMO', plan_expires_at = NOW() + INTERVAL '7 days'`
	_, err := d.db.Exec(query, phone, name, grade, state, district, dialect)
	return err
}

func (d *DemoService) CheckDemoActive(expiresAt time.Time) bool {
	return time.Now().Before(expiresAt)
}
