package temporal

import (
	"fmt"
	"sync"
	"time"
)

type TimeSnapshot struct {
	NowIST               time.Time
	CurrentDateString    string
	FormattedTimestamp   string
	AcademicSessionLabel string
	NextDayDateString    string
	IsMonthLastDay       bool
	IsYearLastDay        bool
	IsSunday             bool
}

type ClockEngine struct {
	loc *time.Location
	mu  sync.RWMutex
}

var GlobalClock *ClockEngine

func init() {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+1800)
	}
	GlobalClock = &ClockEngine{loc: loc}
}

func NewClockEngine() *ClockEngine {
	return GlobalClock
}

func (c *ClockEngine) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Now().In(c.loc)
}

func (c *ClockEngine) GetCurrentSnapshot() TimeSnapshot {
	now := c.Now()
	tomorrow := now.AddDate(0, 0, 1)

	return TimeSnapshot{
		NowIST:               now,
		CurrentDateString:    now.Format("2006-01-02"),
		FormattedTimestamp:   now.Format("02 Jan 2006, 03:04 PM"),
		AcademicSessionLabel: fmt.Sprintf("%d-%d", now.Year(), now.Year()+1),
		NextDayDateString:    tomorrow.Format("2006-01-02"),
		IsMonthLastDay:       now.Month() != tomorrow.Month(),
		IsYearLastDay:        now.Year() != tomorrow.Year(),
		IsSunday:             now.Weekday() == time.Sunday,
	}
}

func (c *ClockEngine) HasDailyResetOccurred(lastReset time.Time) bool {
	now := c.Now()
	lastResetIST := lastReset.In(c.loc)
	return now.Year() != lastResetIST.Year() || now.YearDay() != lastResetIST.YearDay()
}
