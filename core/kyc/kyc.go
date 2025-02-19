package kyc

import "time"

type KYC struct {
	OrgId  string
	Status bool
	Link   string

	CreatedAt time.Time
	UpdatedAt time.Time
}
