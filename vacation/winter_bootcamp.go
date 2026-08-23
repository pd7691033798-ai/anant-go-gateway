package vacation

import (
	"database/sql"
	"fmt"
	"time"
)

type DynamicWinterInfo struct {
	State          string
	District       string
	IsWinterBreak  bool
	DaysLeft       int
	InstructionMsg string
}

type WinterBootcampService struct {
	db *sql.DB
}

func NewWinterBootcampService(db *sql.DB) *WinterBootcampService {
	return &WinterBootcampService{db: db}
}

func (w *WinterBootcampService) ResolveWinterBootcamp(state, district string) DynamicWinterInfo {
	now := time.Now()
	currentYear := now.Year()
	today := now.Format("2006-01-02")
	var holidayType string
	var endDate time.Time

	query := `SELECT holiday_type, end_date FROM state_academic_calendars 
	          WHERE state = $1 AND (district = $2 OR district = 'ALL') 
	            AND holiday_type IN ('WINTER_VACATION', 'COLD_WAVE_DM_ORDER')
	            AND $3 BETWEEN start_date AND end_date AND is_active = TRUE 
	          ORDER BY (district = $2) DESC LIMIT 1`

	err := w.db.QueryRow(query, state, district, today).Scan(&holidayType, &endDate)

	info := DynamicWinterInfo{
		State:    state,
		District: district,
	}

	if err == nil {
		daysLeft := int(endDate.Sub(now).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}
		info.IsWinterBreak = true
		info.DaysLeft = daysLeft
		info.InstructionMsg = fmt.Sprintf("=== विंटर स्पीड बूटकैंप (%s - %s | सत्र %d-%d | शेष %d दिन) ===\n"+
			"1. शीतकालीन अवकाश सक्रिय है। 15-मिनट स्पीड राइटिंग और सामान्य गणितीय गलतियों का सघन सुधार कराएं।",
			state, district, currentYear, currentYear+1, daysLeft)
		return info
	}

	info.IsWinterBreak = false
	info.DaysLeft = 0
	info.InstructionMsg = "नियमित अध्ययन सत्र सक्रिय।"
	return info
}
