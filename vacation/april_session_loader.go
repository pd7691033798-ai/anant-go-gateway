package vacation

import (
	"database/sql"
	"fmt"
	"time"
)

type DynamicSessionInfo struct {
	State          string
	District       string
	IsNewSession   bool
	SessionPhase   string
	InstructionMsg string
}

type AprilSessionService struct {
	db *sql.DB
}

func NewAprilSessionService(db *sql.DB) *AprilSessionService {
	return &AprilSessionService{db: db}
}

func (a *AprilSessionService) ResolveStateAcademicSession(state, district string) DynamicSessionInfo {
	today := time.Now().Format("2006-01-02")
	var holidayType string

	query := `SELECT holiday_type FROM state_academic_calendars 
	          WHERE state = $1 AND (district = $2 OR district = 'ALL') 
	            AND $3 BETWEEN start_date AND end_date AND is_active = TRUE LIMIT 1`

	err := a.db.QueryRow(query, state, district, today).Scan(&holidayType)

	now := time.Now()
	month := int(now.Month())
	day := now.Day()

	info := DynamicSessionInfo{
		State:    state,
		District: district,
	}

	if err == nil && holidayType == "SUMMER_VACATION" {
		info.IsNewSession = false
		info.SessionPhase = "SUMMER_BREAK"
		info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | ग्रीष्मकालीन अवकाश सक्रिय है। दैनिक 15-मिनट हॉलिडे होमवर्क डीकंपोज़र चालू रखें।", state, district)
		return info
	}

	switch state {
	case "Rajasthan", "Haryana", "Punjab", "Delhi", "Uttar Pradesh", "Bihar":
		if month == 4 || (month == 5 && day <= 16) {
			info.IsNewSession = true
			info.SessionPhase = "NEW_SESSION_NORTH_INDIA"
			info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | नया शैक्षणिक सत्र लोड मोड। केवल स्कूल होमवर्क की 15-मिनट स्टेप-मार्किंग करें।", state, district)
			return info
		}
	case "Maharashtra", "Gujarat", "Karnataka", "Tamil Nadu", "Kerala":
		if month == 6 || (month == 7 && day <= 15) {
			info.IsNewSession = true
			info.SessionPhase = "NEW_SESSION_WEST_SOUTH_INDIA"
			info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | नया शैक्षणिक सत्र (जून-जुलाई)। बुनियादी सिद्धांतों और स्कूल होमवर्क की जांच करें।", state, district)
			return info
		}
	}

	info.IsNewSession = false
	info.SessionPhase = "REGULAR_ACADEMIC_STUDY"
	info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | नियमित अध्ययन सत्र सक्रिय। 15-मिनट दैनिक अभ्यास जारी रखें।", state, district)
	return info
}

