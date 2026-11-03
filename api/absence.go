package api

import (
	"fmt"
	"slices"
	"time"
)

// Absence is a recurring time window the vehicle is expected to be away from home
type Absence struct {
	Weekdays []int  `json:"weekdays"` // 0-6 (Sunday-Saturday), day the absence starts
	From     string `json:"from"`     // HH:MM
	To       string `json:"to"`       // HH:MM, before From the absence ends on the following day
	Tz       string `json:"tz"`       // timezone in IANA format
	Active   bool   `json:"active"`   // active flag
}

// Validate checks weekdays, times and timezone
func (a Absence) Validate() error {
	for _, day := range a.Weekdays {
		if day < 0 || day > 6 {
			return fmt.Errorf("weekday out of range: %v", day)
		}
	}
	if _, err := time.LoadLocation(a.Tz); err != nil {
		return fmt.Errorf("invalid timezone: %v", err)
	}
	for _, s := range []string{a.From, a.To} {
		if _, err := time.Parse("15:04", s); err != nil {
			return fmt.Errorf("invalid time: %v", err)
		}
	}
	if a.From == a.To {
		return fmt.Errorf("empty absence: %s-%s", a.From, a.To)
	}
	return nil
}

// Covers returns true if the absence includes the given time
func (a Absence) Covers(t time.Time) bool {
	if !a.Active {
		return false
	}

	loc, err := time.LoadLocation(a.Tz)
	if err != nil {
		return false
	}
	from, err1 := time.Parse("15:04", a.From)
	to, err2 := time.Parse("15:04", a.To)
	if err1 != nil || err2 != nil {
		return false
	}

	lt := t.In(loc)
	minute := lt.Hour()*60 + lt.Minute()
	start := from.Hour()*60 + from.Minute()
	end := to.Hour()*60 + to.Minute()

	startsOn := func(day time.Time) bool {
		return slices.Contains(a.Weekdays, int(day.Weekday()))
	}

	if start < end {
		return startsOn(lt) && minute >= start && minute < end
	}

	// spans midnight
	return startsOn(lt) && minute >= start || startsOn(lt.AddDate(0, 0, -1)) && minute < end
}
