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
	Zone           string `json:"zone"`
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

// GetPanIndiaZone राज्य के आधार पर भारत के 4 प्रमुख भौगोलिक व शैक्षणिक ज़ोन की पहचान करता है
func (w *WinterBootcampService) GetPanIndiaZone(state string) string {
	cleanState := strings.ToLower(strings.TrimSpace(state))

	switch cleanState {
	case "jammu and kashmir", "ladakh", "himachal pradesh", "sikkim", "arunachal pradesh":
		return "HIMALAYAN_ZONE"
	case "tamil nadu", "kerala", "karnataka", "andhra pradesh", "telangana", "puducherry":
		return "SOUTH_ZONE"
	case "maharashtra", "gujarat", "goa":
		return "WEST_ZONE"
	case "rajasthan", "haryana", "punjab", "delhi", "uttar pradesh", "bihar", "madhya pradesh", "uttarakhand", "jharkhand", "chhattisgarh", "chandigarh":
		return "NORTH_PLAINS_ZONE"
	default:
		return "PAN_INDIA_GENERAL"
	}
}

// ResolveWinterBootcamp पूरे भारत (Pan-India) के किसी भी राज्य/जिले के लिए विंटर बूटकैंप स्टेटस निर्धारित करता है
func (w *WinterBootcampService) ResolveWinterBootcamp(ctx context.Context, state, district string) DynamicWinterInfo {
	cleanState := strings.TrimSpace(state)
	if cleanState == "" {
		cleanState = "All India"
	}

	cleanDistrict := strings.TrimSpace(district)
	if cleanDistrict == "" {
		cleanDistrict = "General"
	}

	now := time.Now().In(w.loc)
	currentYear := now.Year()
	currentMonth := int(now.Month())
	currentDay := now.Day()
	todayStr := now.Format("2006-01-02")

	// भारतीय शैक्षणिक सत्र की गणना (1 अप्रैल से 31 मार्च)
	sessionStartYear := currentYear
	sessionEndYear := currentYear + 1
	if currentMonth < 4 {
		sessionStartYear = currentYear - 1
		sessionEndYear = currentYear
	}

	zone := w.GetPanIndiaZone(cleanState)

	info := DynamicWinterInfo{
		Zone:          zone,
		State:         cleanState,
		District:      cleanDistrict,
		IsWinterBreak: false,
		DaysLeft:      0,
	}

	// 1. डेटाबेस से लाइव सरकारी आदेश / डीएम शीतलहर आदेश की प्राथमिक जाँच
	if w.db != nil {
		dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		var holidayType string
		var endDate time.Time

		query := `
			SELECT holiday_type, end_date 
			FROM state_academic_calendars 
			WHERE (state = $1 OR state = 'ALL') 
			  AND (district = $2 OR district = 'ALL') 
			  AND holiday_type IN ('WINTER_VACATION', 'COLD_WAVE_DM_ORDER')
			  AND $3 BETWEEN start_date AND end_date 
			  AND is_active = TRUE 
			ORDER BY (district = $2) DESC, (state = $1) DESC 
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
				"=== पैन-इंडिया विंटर स्पीड बूटकैंप (%s | %s - %s | सत्र %d-%d | शेष %d दिन) ===\n"+
					"1. शीतकालीन अवकाश / डीएम शीतलहर आदेश सक्रिय है।\n"+
					"2. दैनिक 15-मिनट स्पीड राइटिंग, फॉर्मूला रिवीजन और सिली मिस्टेक्स का सघन सुधार कराएं।\n"+
					"3. बच्चों को ठंड के समय छोटे, केंद्रित और व्यावहारिक प्रश्नों से जोड़ें।",
				zone, cleanState, cleanDistrict, sessionStartYear, sessionEndYear, daysLeft,
			)
			return info
		} else if !errors.Is(err, sql.ErrNoRows) {
			// आवश्यकतानुसार लॉगिंग
		}
	}

	// 2. पैन-इंडिया ज़ोनल फ़ॉलबैक (यदि डेटाबेस में विशिष्ट डीएम आदेश न मिले)
	isBreakActive := false
	var estimatedDaysLeft int

	switch zone {
	case "HIMALAYAN_ZONE":
		// हिमालयी क्षेत्र: 15 दिसंबर से 28 फरवरी
		if (currentMonth == 12 && currentDay >= 15) || currentMonth == 1 || currentMonth == 2 {
			isBreakActive = true
			estimatedDaysLeft = 15
		}

	case "NORTH_PLAINS_ZONE":
		// उत्तर भारत: 25 दिसंबर से 15 जनवरी (कड़ाके की ठंड)
		if (currentMonth == 12 && currentDay >= 25) || (currentMonth == 1 && currentDay <= 15) {
			isBreakActive = true
			estimatedDaysLeft = 7
		}

	case "SOUTH_ZONE":
		// दक्षिण भारत: संक्रांति/पोंगल ब्रेक (10 से 18 जनवरी)
		if currentMonth == 1 && (currentDay >= 10 && currentDay <= 18) {
			isBreakActive = true
			estimatedDaysLeft = 5
		}

	case "WEST_ZONE":
		// पश्चिम भारत: 24 दिसंबर से 1 जनवरी
		if (currentMonth == 12 && currentDay >= 24) || (currentMonth == 1 && currentDay <= 1) {
			isBreakActive = true
			estimatedDaysLeft = 4
		}

	default:
		// सामान्य अखिल भारतीय डिफ़ॉल्ट
		if (currentMonth == 12 && currentDay >= 25) || (currentMonth == 1 && currentDay <= 5) {
			isBreakActive = true
			estimatedDaysLeft = 5
		}
	}

	if isBreakActive {
		info.IsWinterBreak = true
		info.DaysLeft = estimatedDaysLeft
		info.InstructionMsg = fmt.Sprintf(
			"=== पैन-इंडिया विंटर स्पीड बूटकैंप (%s | %s | सत्र %d-%d) ===\n"+
				"1. क्षेत्रीय शीतकालीन अवकाश सक्रिय है।\n"+
				"2. दैनिक 15-मिनट फॉर्मूला रिवीजन और हैंडराइटिंग स्पीड टेस्ट चालू रखें।",
			zone, cleanState, sessionStartYear, sessionEndYear,
		)
		return info
	}

	// 3. डिफ़ॉल्ट: नियमित अध्ययन सत्र
	info.IsWinterBreak = false
	info.DaysLeft = 0
	info.InstructionMsg = fmt.Sprintf("पैन-इंडिया नियमित अध्ययन सत्र सक्रिय (%s, %s)। 15-मिनट दैनिक अभ्यास जारी रखें।", cleanState, cleanDistrict)
	return info
}
