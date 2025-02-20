package kyc

import "time"

type KYC struct {
	OrgID  string
	Status bool
	Link   string

	CreatedAt time.Time
	UpdatedAt time.Time
}
