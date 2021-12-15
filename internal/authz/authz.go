package authz

import (
	"context"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/authz/spicedb"
)

type Policy interface {
	AddPolicy(ctx context.Context, schema string) error
}

type Authz struct {
	Policy
}

func New(config *config.Shield, logger log.Logger) *Authz {
	spice, err := spicedb.New(config.SpiceDB, logger)

	if err != nil {
		logger.Fatal(err.Error())
	}

	return &Authz{
		spice.Policy,
	}
}
