package temporal

import (
	"fmt"
	"sync"
	"time"
)

type TimeSnapshot struct {
	NowIST               time.Time     `json:"now_ist"`
	CurrentDateString    string        `json:"current_date_string"`
	FormattedTimestamp   string        `json:"formatted_timestamp"`
	AcademicSessionLabel string        `json:"academic_session_label"`
	NextDayDateString    string        `json:"next_day_date_string"`
	IsMonthLastDay       bool          `json:"is_month_last_day"`
	IsYearLastDay        bool          `json:"is_year_last_day"`
	IsSunday             bool          `json:"is_sunday"`
	DurationToMidnight   time.Duration `json:"duration_to_midnight"`
}

type MasterClockEngine struct {
	loc *time.Location
	mu  sync.RWMutex
}

var GlobalMasterClock *MasterClockEngine

func init() {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+1800)
	}
	GlobalMasterClock = &MasterClockEngine{loc: loc}
}

func GetMasterClock() *MasterClockEngine {
	return GlobalMasterClock
}

// Now वर्तमान समय को IST में थ्रेड-सेफ़ तरीके से लौटाता है
func (c *MasterClockEngine) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Now().In(c.loc)
}

// GetCurrentSnapshot सिस्टम-वाइड टाइम और कैलेंडर का संपूर्ण स्नैपशॉट देता है
func (c *MasterClockEngine) GetCurrentSnapshot() TimeSnapshot {
	now := c.Now()
	year, month, day := now.Date()
	tomorrow := now.AddDate(0, 0, 1)

	// भारतीय शैक्षणिक सत्र की गणना (1 अप्रैल से 31 मार्च)
	sessionStartYear := year
	sessionEndYear := year + 1
	if int(month) < 4 {
		sessionStartYear = year - 1
		sessionEndYear = year
	}
	sessionLabel := fmt.Sprintf("%d-%d", sessionStartYear, sessionEndYear)

	// मध्यरात्रि तक का समय
	nextMidnight := time.Date(year, month, day+1, 0, 0, 0, 0, c.loc)
	durationToMidnight := nextMidnight.Sub(now)

	return TimeSnapshot{
		NowIST:               now,
		CurrentDateString:    now.Format("2006-01-02"),
		FormattedTimestamp:   now.Format("02 Jan 2006, 03:04 PM"),
		AcademicSessionLabel: sessionLabel,
		NextDayDateString:    tomorrow.Format("2006-01-02"),
		IsMonthLastDay:       now.Month() != tomorrow.Month(),
		IsYearLastDay:        now.Year() != tomorrow.Year(),
		IsSunday:             now.Weekday() == time.Sunday,
		DurationToMidnight:   durationToMidnight,
	}
}

// HasDailyResetOccurred जाँचता है कि क्या अंतिम रीसेट के बाद नई तारीख शुरू हो चुकी है
func (c *MasterClockEngine) HasDailyResetOccurred(lastReset time.Time) bool {
	now := c.Now()
	lastResetIST := lastReset.In(c.loc)
	return now.Year() != lastResetIST.Year() || now.YearDay() != lastResetIST.YearDay()
}
