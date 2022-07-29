package project

import (
	"context"
	"strings"
	"time"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/metadata"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Project, error)
	GetBySlug(ctx context.Context, slug string) (Project, error)
	Create(ctx context.Context, org Project) (Project, error)
	List(ctx context.Context) ([]Project, error)
	UpdateByID(ctx context.Context, toUpdate Project) (Project, error)
	UpdateBySlug(ctx context.Context, toUpdate Project) (Project, error)
	ListAdmins(ctx context.Context, id string) ([]user.User, error)
}

type Project struct {
	ID           string
	Name         string
	Slug         string
	Organization organization.Organization
	Metadata     metadata.Metadata
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (p Project) GenerateSlug() string {
	if strings.TrimSpace(p.Slug) != "" {
		return p.Slug
	}
	preProcessed := strings.ReplaceAll(strings.TrimSpace(strings.TrimSpace(p.Name)), "_", "-")
	return strings.Join(
		strings.Split(preProcessed, " "),
		"-",
	)
}
