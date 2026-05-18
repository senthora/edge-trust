package testutil

import "time"

func FixedTimeNow(location *time.Location) time.Time {
	return time.Date(2026, 5, 6, 14, 30, 0, 0, location)
}

func FixedTimeNowUTC() time.Time {
	return time.Date(2026, 5, 6, 14, 30, 0, 0, time.UTC)
}
