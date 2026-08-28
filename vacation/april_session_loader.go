package vacation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type DynamicSessionInfo struct {
	State          string `json:"state"`
	District       string `json:"district"`
	IsNewSession   bool   `json:"is_new_session"`
	IsHoliday      bool   `json:"is_holiday"`
	SessionPhase   string `json:"session_phase"`
	InstructionMsg string `json:"instruction_msg"`
}

type AprilSessionService struct {
	db  *sql.DB
	loc *time.Location
}

func NewAprilSessionService(db *sql.DB) *AprilSessionService {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	return &AprilSessionService{
		db:  db,
		loc: loc,
	}
}

// ResolveStateAcademicSession राज्य, जिला और वर्तमान तिथि के आधार पर सक्रिय शैक्षणिक सत्र या अवकाश का निर्धारण करता है
func (a *AprilSessionService) ResolveStateAcademicSession(ctx context.Context, state, district string) DynamicSessionInfo {
	cleanState := strings.TrimSpace(state)
	if cleanState == "" {
		cleanState = "Rajasthan"
	}

	cleanDistrict := strings.TrimSpace(district)
	if cleanDistrict == "" {
		cleanDistrict = "Sri Ganganagar"
	}

	now := time.Now().In(a.loc)
	todayStr := now.Format("2006-01-02")
	month := int(now.Month())
	day := now.Day()

	info := DynamicSessionInfo{
		State:    cleanState,
		District: cleanDistrict,
	}

	// 1. डेटाबेस से सक्रिय अवकाश (Holiday/Break) की जाँच
	if a.db != nil {
		dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		var holidayType string
		query := `
			SELECT holiday_type 
			FROM state_academic_calendars 
			WHERE state = $1 
			  AND (district = $2 OR district = 'ALL') 
			  AND $3 BETWEEN start_date AND end_date 
			  AND is_active = TRUE 
			LIMIT 1`

		err := a.db.QueryRowContext(dbCtx, query, cleanState, cleanDistrict, todayStr).Scan(&holidayType)
		if err == nil {
			info.IsHoliday = true
			if holidayType == "SUMMER_VACATION" {
				info.IsNewSession = false
				info.SessionPhase = "SUMMER_BREAK"
				info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | ग्रीष्मकालीन अवकाश सक्रिय है। दैनिक 15-मिनट हॉलिडे होमवर्क डीकंपोज़र चालू रखें।", cleanState, cleanDistrict)
				return info
			} else if holidayType == "WINTER_VACATION" {
				info.IsNewSession = false
				info.SessionPhase = "WINTER_BREAK"
				info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | शीतकालीन अवकाश सक्रिय है। दैनिक रिवीज़न और वार्म-अप अभ्यास जारी रखें।", cleanState, cleanDistrict)
				return info
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			// लॉगिंग की जा सकती है यदि आवश्यक हो
		}
	}

	// 2. क्षेत्रीय कैलेंडर नियमों के आधार पर नया सत्र निर्धारण
	switch cleanState {
	case "Rajasthan", "Haryana", "Punjab", "Delhi", "Uttar Pradesh", "Bihar", "Madhya Pradesh", "Uttarakhand", "Himachal Pradesh":
		// उत्तर और मध्य भारत: नया सत्र 1 अप्रैल से मई मध्य तक
		if month == 4 || (month == 5 && day <= 16) {
			info.IsNewSession = true
			info.IsHoliday = false
			info.SessionPhase = "NEW_SESSION_NORTH_INDIA"
			info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | नया शैक्षणिक सत्र (अप्रैल-मई)। केवल स्कूल होमवर्क की 15-मिनट स्टेप-मार्किंग करें।", cleanState, cleanDistrict)
			return info
		}

	case "Maharashtra", "Gujarat", "Karnataka", "Tamil Nadu", "Kerala", "Andhra Pradesh", "Telangana":
		// पश्चिम और दक्षिण भारत: नया सत्र जून से जुलाई मध्य तक
		if month == 6 || (month == 7 && day <= 15) {
			info.IsNewSession = true
			info.IsHoliday = false
			info.SessionPhase = "NEW_SESSION_WEST_SOUTH_INDIA"
			info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | नया शैक्षणिक सत्र (जून-जुलाई)। बुनियादी सिद्धांतों और स्कूल होमवर्क की जांच करें।", cleanState, cleanDistrict)
			return info
		}
	}

	// 3. डिफ़ॉल्ट: नियमित अध्ययन सत्र
	info.IsNewSession = false
	info.IsHoliday = false
	info.SessionPhase = "REGULAR_ACADEMIC_STUDY"
	info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | नियमित अध्ययन सत्र सक्रिय। 15-मिनट दैनिक अभ्यास जारी रखें।", cleanState, cleanDistrict)
	return info
}

