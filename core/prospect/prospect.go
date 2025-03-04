package prospect

import (
	"context"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type Status string

const (
	Unsubscribed Status = "unsubscribed"
	Subscribed   Status = "subscribed"
)

func (s Status) String() string {
	return string(s)
}

func StringToStatus(s string) Status {
	switch s {
	case "status_unsubscribed":
		return Unsubscribed
	case "status_subscribed":
		return Subscribed
	default:
		return Unsubscribed
	}
}

func (s Status) ToDB() Status {
	return s
}

type Prospect struct {
	ID        string
	Name      string
	Email     string
	Phone     string
	Activity  string
	Status    Status // subscription status
	ChangedAt time.Time
	Source    string
	Verified  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Metadata  metadata.Metadata
}

type Repository interface {
	Create(ctx context.Context, prospect Prospect) (Prospect, error)
	Get(ctx context.Context, id string) (Prospect, error)
	List(ctx context.Context, filter Filter) ([]Prospect, error)
	Update(ctx context.Context, prospect Prospect) (Prospect, error)
	Delete(ctx context.Context, id string) error
}
