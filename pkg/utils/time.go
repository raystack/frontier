package utils

import "time"

func AsTimeFromEpoch(unixEpoch int64) time.Time {
	if unixEpoch == 0 {
		return time.Time{}
	}
	return time.Unix(unixEpoch, 0)
}

func ConvertToStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
