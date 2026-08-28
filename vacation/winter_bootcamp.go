package vacation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type DynamicWinterInfo struct {
	State          string `json:"state"`
	District       string `json:"district"`
	IsWinterBreak  bool   `json:"is_winter_break"`
	DaysLeft       int    `json:"days_left"`
	InstructionMsg string `json:"instruction_msg"`
}

type WinterBootcampService struct {
	db  *sql.DB
	loc *time.Location
}

func NewWinterBootcampService(db *sql.DB) *WinterBootcampService {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	return &WinterBootcampService{
		db:  db,
		loc: loc,
	}
}

// ResolveWinterBootcamp शीतकालीन अवकाश और शीतलहर आदेशों की पुष्टि कर बूटकैंप निर्देश तैयार करता है
func (w *WinterBootcampService) ResolveWinterBootcamp(ctx context.Context, state, district string) DynamicWinterInfo {
	cleanState := strings.TrimSpace(state)
	if cleanState == "" {
		cleanState = "Rajasthan"
	}

	cleanDistrict := strings.TrimSpace(district)
	if cleanDistrict == "" {
		cleanDistrict = "Sri Ganganagar"
	}

	now := time.Now().In(w.loc)
	currentYear := now.Year()
	currentMonth := int(now.Month())
	todayStr := now.Format("2006-01-02")

	// शैक्षणिक सत्र लेबल की सटीक गणना (1 अप्रैल से 31 मार्च)
	sessionStartYear := currentYear
	sessionEndYear := currentYear + 1
	if currentMonth < 4 {
		sessionStartYear = currentYear - 1
		sessionEndYear = currentYear
	}

	info := DynamicWinterInfo{
		State:         cleanState,
		District:      cleanDistrict,
		IsWinterBreak: false,
		DaysLeft:      0,
	}

	// 1. डेटाबेस से शीतकालीन अवकाश / डीएम शीतलहर आदेश की जाँच
	if w.db != nil {
		dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		var holidayType string
		var endDate time.Time

		query := `
			SELECT holiday_type, end_date 
			FROM state_academic_calendars 
			WHERE state = $1 
			  AND (district = $2 OR district = 'ALL') 
			  AND holiday_type IN ('WINTER_VACATION', 'COLD_WAVE_DM_ORDER')
			  AND $3 BETWEEN start_date AND end_date 
			  AND is_active = TRUE 
			ORDER BY (district = $2) DESC 
			LIMIT 1`

		err := w.db.QueryRowContext(dbCtx, query, cleanState, cleanDistrict, todayStr).Scan(&holidayType, &endDate)
		if err == nil {
			daysLeft := int(endDate.In(w.loc).Sub(now).Hours() / 24)
			if daysLeft < 0 {
				daysLeft = 0
			}

			info.IsWinterBreak = true
			info.DaysLeft = daysLeft
			info.InstructionMsg = fmt.Sprintf(
				"=== विंटर स्पीड बूटकैंप (%s - %s | सत्र %d-%d | शेष %d दिन) ===\n"+
					"1. शीतकालीन अवकाश / शीतलहर अवकाश सक्रिय है।\n"+
					"2. दैनिक 15-मिनट स्पीड राइटिंग, फॉर्मूला रिवीजन और सिली मिस्टेक्स का सघन सुधार कराएं।\n"+
					"3. बच्चों को ठंड के समय छोटे, केंद्रित और हाथ से लिखने वाले व्यावहारिक प्रश्नों से जोड़ें।",
				cleanState, cleanDistrict, sessionStartYear, sessionEndYear, daysLeft,
			)
			return info
		} else if !errors.Is(err, sql.ErrNoRows) {
			// लॉगिंग की जा सकती है यदि आवश्यक हो
		}
	}

	// 2. क्षेत्रीय फ़ॉलबैक नियम (यदि डेटाबेस में विशेष तिथि न मिली हो)
	decomposer := NewHomeworkDecomposerService(w.db)
	if decomposer.IsWinterBreak(cleanState) {
		info.IsWinterBreak = true
		info.DaysLeft = 7 // अनुमानित शेष दिन
		info.InstructionMsg = fmt.Sprintf(
			"=== विंटर स्पीड बूटकैंप (%s | सत्र %d-%d) ===\n"+
				"1. क्षेत्रीय शीतकालीन अवकाश सक्रिय है।\n"+
				"2. दैनिक 15-मिनट रिवीज़न और गणितीय सूत्रों का हस्तलेखन अभ्यास जारी रखें।",
			cleanState, sessionStartYear, sessionEndYear,
		)
		return info
	}

	// 3. डिफ़ॉल्ट: नियमित अध्ययन सत्र
	info.IsWinterBreak = false
	info.DaysLeft = 0
	info.InstructionMsg = fmt.Sprintf("राज्य: %s, ज़िला: %s | नियमित अध्ययन सत्र सक्रिय।", cleanState, cleanDistrict)
	return info
}

