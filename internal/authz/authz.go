package authz

import (
	"context"

	"github.com/odpf/shield/model"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/authz/spicedb"
)

type Policy interface {
	AddPolicy(ctx context.Context, schema string) error
}

type Permission interface {
	AddRelation(ctx context.Context, relation model.Relation) error
	DeleteRelation(ctx context.Context, relation model.Relation) error
	CheckRelation(ctx context.Context, relation model.Relation, permissionID model.Permission) (bool, error)
}

type Authz struct {
	Policy
	Permission
}

func New(config *config.Shield, logger log.Logger) *Authz {
	spice, err := spicedb.New(config.SpiceDB, logger)

	if err != nil {
		logger.Fatal(err.Error())
	}

	return &Authz{
		spice.Policy,
		spice.Permission,
	}
}
