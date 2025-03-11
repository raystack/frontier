package prospect

import (
	"context"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/salt/rql"
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
	ID        string    `rql:"name=id,type=string"`
	Name      string    `rql:"name=name,type=string"`
	Email     string    `rql:"name=email,type=string"`
	Phone     string    `rql:"name=phone,type=string"`
	Activity  string    `rql:"name=activity,type=string"`
	Status    Status    `rql:"name=status,type=string"` // subscription status
	ChangedAt time.Time `rql:"name=changed_at,type=datetime"`
	Source    string    `rql:"name=source,type=string"`
	Verified  bool      `rql:"name=verified,type=bool"`
	CreatedAt time.Time `rql:"name=created_at,type=datetime"`
	UpdatedAt time.Time `rql:"name=updated_at,type=datetime"`
	Metadata  metadata.Metadata
}

type Repository interface {
	Create(ctx context.Context, prospect Prospect) (Prospect, error)
	Get(ctx context.Context, id string) (Prospect, error)
	List(ctx context.Context, query *rql.Query) ([]Prospect, error)
	Update(ctx context.Context, prospect Prospect) (Prospect, error)
	Delete(ctx context.Context, id string) error
}
