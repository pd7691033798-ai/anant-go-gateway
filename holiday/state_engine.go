package holiday

import (
	"database/sql"
	"strings"
	"time"
)

type StateHolidayService struct {
	db *sql.DB
}

func NewStateHolidayService(db *sql.DB) *StateHolidayService {
	return &StateHolidayService{db: db}
}

// ParseDateFlexible: यह फंक्शन खुद पहचान लेगा कि तारीख किस फॉर्मेट में है
func ParseDateFlexible(dateStr string) (time.Time, bool) {
	dateStr = strings.TrimSpace(dateStr)
	
	// संभावित फॉर्मेट्स जिनकी भारत में या डेटाबेस में सबसे ज्यादा उम्मीद होती है
	layouts := []string{
		"02-01-2006", // DD-MM-YYYY (भारतीय मानक)
		"2006-01-02", // YYYY-MM-DD (डेटाबेस स्टैंडर्ड)
		"02/01/2006", // DD/MM/YYYY
		"2006/01/02", // YYYY/MM/DD
		"02-Jan-2006", // 22-Sep-2026 जैसे फॉर्मेट
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func (s *StateHolidayService) CheckHoliday(state, district string) (bool, string, int) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	todayStr := now.Format("02-01-2006")

	var holidayType string
	var endDateStr string

	// डेटाबेस क्वेरी जो किसी भी फॉर्मेट की परवाह किए बिना सही रेंज पकड़े
	query := `SELECT holiday_type, TO_CHAR(end_date, 'DD-MM-YYYY') FROM state_academic_calendars 
	          WHERE state = $1 AND (district = $2 OR district = 'ALL') 
	            AND TO_DATE($3, 'DD-MM-YYYY') BETWEEN start_date AND end_date AND is_active = TRUE 
	          ORDER BY CASE WHEN district = $2 THEN 1 ELSE 2 END LIMIT 1`

	err = s.db.QueryRow(query, state, district, todayStr).Scan(&holidayType, &endDateStr)
	if err != nil {
		return false, "REGULAR_SCHOOL_DAY", 0
	}

	// फ्लेक्सिबल पार्सर का इस्तेमाल करके छुट्टी की आखिरी तारीख निकालना
	endDate, ok := ParseDateFlexible(endDateStr)
	if !ok {
		return false, "REGULAR_SCHOOL_DAY", 0
	}

	todayDate, _ := time.Parse("02-01-2006", todayStr)
	daysLeft := int(endDate.Sub(todayDate).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}

	return true, holidayType, daysLeft
}
