package bootstrap

import (
	"context"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/internal/authz"
	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/model"
)

// Insert Action
// Insert Policy

type Store interface {
	CreateNamespace(ctx context.Context, namespace model.Namespace) (model.Namespace, error)
	CreateRole(ctx context.Context, role model.Role) (model.Role, error)
}

func BootstrapDefinitions(ctx context.Context, store Store, authz *authz.Authz, logger log.Logger) {
	bootstrapNamespaces(ctx, store, logger)
	bootstrapRoles(ctx, store, logger)
}

func bootstrapRoles(ctx context.Context, store Store, logger log.Logger) {
	roles := []model.Role{
		definition.OrganizationAdminRole,
		definition.ProjectAdminRole,
		definition.TeamAdminRole,
		definition.TeamMemberRole,
	}

	for _, role := range roles {
		_, err := store.CreateRole(ctx, role)
		if err != nil {
			logger.Fatal(err.Error())
		}
	}

	logger.Info("Bootstrap Roles Successfully")
}

func bootstrapNamespaces(ctx context.Context, store Store, logger log.Logger) {
	namespaces := []model.Namespace{
		definition.OrgNamespace,
		definition.ProjectNamespace,
		definition.TeamNamespace,
		definition.UserNamespace,
	}

	for _, ns := range namespaces {
		_, err := store.CreateNamespace(ctx, ns)
		if err != nil {
			logger.Fatal(err.Error())
		}
	}
	logger.Info("Bootstrap Namespaces Successfully")
}
