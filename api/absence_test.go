package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAbsenceCovers(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Berlin")
	at := func(day, hour, minute int) time.Time {
		// 2026-11-02 is a Monday
		return time.Date(2026, 11, 2+day, hour, minute, 0, 0, loc)
	}

	work := Absence{Weekdays: []int{1, 2, 3, 4, 5}, From: "05:00", To: "16:00", Tz: "Europe/Berlin", Active: true}

	for _, tc := range []struct {
		t        time.Time
		expected bool
	}{
		{at(0, 4, 59), false},
		{at(0, 5, 0), true},
		{at(0, 15, 59), true},
		{at(0, 16, 0), false},
		{at(5, 10, 0), false}, // Saturday
	} {
		assert.Equal(t, tc.expected, work.Covers(tc.t), tc.t)
	}

	inactive := work
	inactive.Active = false
	assert.False(t, inactive.Covers(at(0, 10, 0)))

	// night shift Friday 22:00 to Saturday 06:00
	night := Absence{Weekdays: []int{5}, From: "22:00", To: "06:00", Tz: "Europe/Berlin", Active: true}

	for _, tc := range []struct {
		t        time.Time
		expected bool
	}{
		{at(4, 21, 59), false},
		{at(4, 22, 0), true},
		{at(5, 5, 59), true},
		{at(5, 6, 0), false},
		{at(5, 22, 0), false}, // Saturday does not start an absence
		{at(0, 3, 0), false},  // Monday after a Sunday without absence
	} {
		assert.Equal(t, tc.expected, night.Covers(tc.t), tc.t)
	}

	// the timezone of the absence applies, not the one of the time
	assert.True(t, work.Covers(at(0, 10, 0).UTC()))
}

func TestAbsenceValidate(t *testing.T) {
	valid := Absence{Weekdays: []int{1}, From: "05:00", To: "16:00", Tz: "Europe/Berlin"}
	assert.NoError(t, valid.Validate())

	for _, a := range []Absence{
		{Weekdays: []int{7}, From: "05:00", To: "16:00", Tz: "Europe/Berlin"},
		{Weekdays: []int{1}, From: "5", To: "16:00", Tz: "Europe/Berlin"},
		{Weekdays: []int{1}, From: "05:00", To: "16:00", Tz: "Mars/Olympus"},
		{Weekdays: []int{1}, From: "05:00", To: "05:00", Tz: "Europe/Berlin"},
	} {
		assert.Error(t, a.Validate(), a)
	}
}
