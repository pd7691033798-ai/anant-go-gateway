package holiday

import (
	"database/sql"
	"time"
)

type StateHolidayService struct {
	db *sql.DB
}

func NewStateHolidayService(db *sql.DB) *StateHolidayService {
	return &StateHolidayService{db: db}
}

func (s *StateHolidayService) CheckHoliday(state, district string) (bool, string, int) {
	now := time.Now()
	today := now.Format("2006-01-02")
	var holidayType string
	var endDate time.Time

	query := `SELECT holiday_type, end_date FROM state_academic_calendars 
	          WHERE state = $1 AND (district = $2 OR district = 'ALL') 
	            AND $3 BETWEEN start_date AND end_date AND is_active = TRUE 
	          ORDER BY (district = $2) DESC LIMIT 1`

	err := s.db.QueryRow(query, state, district, today).Scan(&holidayType, &endDate)
	if err != nil {
		return false, "REGULAR_SCHOOL_DAY", 0
	}

	daysLeft := int(endDate.Sub(now).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}
	return true, holidayType, daysLeft
}
