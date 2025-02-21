package audience

import (
	"context"
	"strconv"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type Status int

const (
	Unsubscribed Status = iota
	Subscribed
)

func (s Status) String() string {
	return strconv.Itoa(int(s))
}

type Audience struct {
	ID        string
	Name      string
	Email     string
	Phone     string
	Activity  string
	Status    Status // subscription status
	ChangedAt *time.Time
	Source    string
	Verified  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Metadata  metadata.Metadata
}

type Repository interface {
	Create(ctx context.Context, audience Audience) (Audience, error)
}
