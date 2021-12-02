package authz

import (
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/authz/spicedb"
)

type Policy interface {
	AddPolicy(schema string) error
}

type Authz struct {
	Policy
}

func New(config *config.Shield, logger log.Logger) *Authz {
	spice, err := spicedb.New(config.SpiceDB)

	if err != nil {
		logger.Fatal(err.Error())
	}
	return &Authz{
		spice.Policy,
	}
}
