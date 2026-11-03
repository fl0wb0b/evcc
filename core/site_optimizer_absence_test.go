package core

import (
	"testing"
	"time"

	"github.com/evcc-io/evcc/api"
	optimizer "github.com/evcc-io/optimizer/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAbsenceAvailability(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Berlin")
	work := []api.Absence{{Weekdays: []int{1, 2, 3, 4, 5}, From: "05:00", To: "16:00", Tz: "Europe/Berlin", Active: true}}

	hourly := func(n int) []int {
		dt := make([]int, n)
		for i := range dt {
			dt[i] = 3600
		}
		return dt
	}

	// Monday 03:00, away from 05:00 to 16:00
	now := time.Date(2026, 11, 2, 3, 0, 0, 0, loc)
	res := absenceAvailability(work, hourly(15), now)
	require.Len(t, res, 15)
	for i, available := range res {
		hour := 3 + i
		assert.Equal(t, hour < 5 || hour >= 16, available, "hour %d", hour)
	}

	// connected during the absence: the ongoing one is not taken, the next day's is
	now = time.Date(2026, 11, 2, 10, 0, 0, 0, loc)
	res = absenceAvailability(work, hourly(24), now)
	for i, available := range res {
		ts := now.Add(time.Duration(i) * time.Hour)
		away := ts.Day() == 3 && ts.Hour() >= 5 && ts.Hour() < 16
		assert.Equal(t, !away, available, ts)
	}

	// never away within the horizon
	now = time.Date(2026, 11, 7, 10, 0, 0, 0, loc) // Saturday
	assert.Nil(t, absenceAvailability(work, hourly(12), now))
	assert.Nil(t, absenceAvailability(nil, hourly(12), now))
}

func TestCoarsenAvailability(t *testing.T) {
	now := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)

	dt := make([]int, 8)
	for i := range dt {
		dt[i] = 900
	}
	req := optimizer.OptimizationInput{
		TimeSeries: optimizer.TimeSeries{Dt: dt, Gt: make([]float32, 8), Ft: make([]float32, 8), PN: make([]float32, 8), PE: make([]float32, 8)},
		Batteries:  []optimizer.BatteryConfig{{CAvailable: []bool{true, true, true, true, true, true, false, true}}},
	}
	details := requestDetails{Timestamps: asTimestamps(dt, now.Add(-time.Second))}

	coarsenHorizon(&req, &details, 0)

	// 23:45, then 00:00 to 01:00 all available, then 01:00 to 01:45 with one slot away
	assert.Equal(t, []bool{true, true, false}, req.Batteries[0].CAvailable)
}
