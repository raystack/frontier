package utils

import "time"

func AsTimeFromEpoch(unixEpoch int64) time.Time {
	if unixEpoch == 0 {
		return time.Time{}
	}
	return time.Unix(unixEpoch, 0)
}
