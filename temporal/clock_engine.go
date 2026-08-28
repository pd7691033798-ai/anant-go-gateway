package temporal

import (
	"fmt"
	"time"
)

// TemporalSnapshot वर्तमान समय, भारतीय शैक्षणिक सत्र एवं कैलेंडर ट्रिगर्स का स्नैपशॉट
type TemporalSnapshot struct {
	CurrentTime            time.Time     `json:"current_time"`
	CurrentYear            int           `json:"current_year"`
	CurrentMonth           int           `json:"current_month"`
	CurrentDay             int           `json:"current_day"`
	CurrentHour            int           `json:"current_hour"`
	FormattedDate          string        `json:"formatted_date"`
	FormattedTimestamp     string        `json:"formatted_timestamp"`
	IsMonthLastDay         bool          `json:"is_month_last_day"`
	IsYearLastDay          bool          `json:"is_year_last_day"`
	IsMonthFirstDay        bool          `json:"is_month_first_day"`
	IsYearFirstDay         bool          `json:"is_year_first_day"`
	IsPeakStudyHour        bool          `json:"is_peak_study_hour"` // शाम 6:00 PM से 9:00 PM
	NextDayDateString      string        `json:"next_day_date_string"`
	AcademicSessionLabel   string        `json:"academic_session_label"`
	DurationToNextMidnight time.Duration `json:"duration_to_next_midnight"`
}

type ClockEngine struct {
	loc *time.Location
}

func NewClockEngine() *ClockEngine {
	// भारतीय मानक समय (Asia/Kolkata - IST UTC+05:30)
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	return &ClockEngine{loc: loc}
}

// GetCurrentSnapshot वर्तमान समय का विस्तृत शैक्षणिक व सुरक्षा स्नैपशॉट उत्पन्न करता है
func (c *ClockEngine) GetCurrentSnapshot() TemporalSnapshot {
	now := time.Now().In(c.loc)
	year, month, day := now.Date()
	hour := now.Hour()

	tomorrow := now.AddDate(0, 0, 1)

	isMonthLastDay := tomorrow.Month() != month
	isYearLastDay := tomorrow.Year() != year
	isMonthFirstDay := (day == 1)
	isYearFirstDay := (month == time.January && day == 1)

	// शाम 6 बजे से रात 9 बजे (18:00 - 21:00 IST) का पीक स्टडी समय
	isPeakStudyHour := (hour >= 18 && hour < 21)

	// भारतीय शैक्षणिक सत्र की गणना (1 अप्रैल से 31 मार्च)
	sessionStartYear := year
	sessionEndYear := year + 1
	if int(month) < 4 {
		sessionStartYear = year - 1
		sessionEndYear = year
	}
	sessionLabel := fmt.Sprintf("सत्र %d-%d", sessionStartYear, sessionEndYear)

	// अगली मध्यरात्रि तक का सटीक समय (दैनिक कोटा और इन-मेमोरी कैश रीसेट के लिए)
	nextMidnight := time.Date(year, month, day+1, 0, 0, 0, 0, c.loc)
	durationToMidnight := nextMidnight.Sub(now)

	return TemporalSnapshot{
		CurrentTime:            now,
		CurrentYear:            year,
		CurrentMonth:           int(month),
		CurrentDay:             day,
		CurrentHour:            hour,
		FormattedDate:          now.Format("2006-01-02"),
		FormattedTimestamp:     now.Format("2006-01-02 15:04:05 MST"),
		IsMonthLastDay:         isMonthLastDay,
		IsYearLastDay:          isYearLastDay,
		IsMonthFirstDay:        isMonthFirstDay,
		IsYearFirstDay:         isYearFirstDay,
		IsPeakStudyHour:        isPeakStudyHour,
		NextDayDateString:      tomorrow.Format("2006-01-02"),
		AcademicSessionLabel:   sessionLabel,
		DurationToNextMidnight: durationToMidnight,
	}
}

// FormatCustomDate किसी भी टाइमस्टैम्प को मानक भारतीय तिथि प्रारूप (DD/MM/YYYY) में बदलता है
func (c *ClockEngine) FormatIndianDate(t time.Time) string {
	return t.In(c.loc).Format("02/01/2006")
}
