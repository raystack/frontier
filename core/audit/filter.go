package audit

import "time"

type Filter struct {
	OrgID  string
	Source string
	Action string

	StartTime time.Time
	EndTime   time.Time

	IgnoreSystem bool
}
