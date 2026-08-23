package temporal

import (
	"fmt"
	"time"
)

type TemporalSnapshot struct {
	CurrentTime          time.Time
	CurrentYear          int
	CurrentMonth         int
	CurrentDay           int
	FormattedDate        string
	FormattedTimestamp   string
	IsMonthLastDay       bool
	IsYearLastDay        bool
	IsMonthFirstDay      bool
	IsYearFirstDay       bool
	NextDayDateString    string
	AcademicSessionLabel string
}

type ClockEngine struct {
	loc *time.Location
}

func NewClockEngine() *ClockEngine {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	return &ClockEngine{loc: loc}
}

func (c *ClockEngine) GetCurrentSnapshot() TemporalSnapshot {
	now := time.Now().In(c.loc)
	year, month, day := now.Date()

	tomorrow := now.AddDate(0, 0, 1)

	isMonthLastDay := tomorrow.Month() != month
	isYearLastDay := tomorrow.Year() != year
	isMonthFirstDay := (day == 1)
	isYearFirstDay := (month == time.January && day == 1)

	sessionStartYear := year
	sessionEndYear := year + 1
	if int(month) < 4 {
		sessionStartYear = year - 1
		sessionEndYear = year
	}
	sessionLabel := fmt.Sprintf("सत्र %d-%d", sessionStartYear, sessionEndYear)

	return TemporalSnapshot{
		CurrentTime:          now,
		CurrentYear:          year,
		CurrentMonth:         int(month),
		CurrentDay:           day,
		FormattedDate:        now.Format("2006-01-02"),
		FormattedTimestamp:   now.Format("2006-01-02 15:04:05 MST"),
		IsMonthLastDay:       isMonthLastDay,
		IsYearLastDay:        isYearLastDay,
		IsMonthFirstDay:      isMonthFirstDay,
		IsYearFirstDay:       isYearFirstDay,
		NextDayDateString:    tomorrow.Format("2006-01-02"),
		AcademicSessionLabel: sessionLabel,
	}
}
