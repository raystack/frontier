package blob

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/resource"

	"github.com/ghodss/yaml"
	"github.com/odpf/salt/log"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"

	"gocloud.dev/blob"
)

type ResourceBackends struct {
	Backends []ResourceBackend `json:"backends" yaml:"backends"`
}

type ResourceBackend struct {
	Name          string         `json:"name" yaml:"name"`
	ResourceTypes []ResourceType `json:"resource_types" yaml:"resource_types"`
}

type ResourceType struct {
	Name    string              `json:"name" yaml:"name"`
	Actions map[string][]string `json:"actions" yaml:"actions"`
}

type Resources struct {
	Resources []Resource
}

type Resource struct {
	Name    string
	Actions map[string][]string
}

type ResourcesRepository struct {
	log log.Logger
	mu  *sync.Mutex

	cron   *cron.Cron
	bucket Bucket
	cached []resource.YAML
}

func (repo *ResourcesRepository) GetAll(ctx context.Context) ([]resource.YAML, error) {
	repo.mu.Lock()
	currentCache := repo.cached
	repo.mu.Unlock()
	if repo.cron != nil {
		// cache must have been refreshed automatically, just return
		return currentCache, nil
	}

	err := repo.refresh(ctx)
	return repo.cached, err
}

func (repo *ResourcesRepository) GetRelationsForNamespace(ctx context.Context, namespaceID string) (map[string]bool, error) {
	resources, err := repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	relationSet := map[string]bool{}
	relationSet["organization"] = true
	relationSet["project"] = true
	relationSet["team"] = true

	for _, resource := range resources {
		if resource.Name == namespaceID {
			for _, action := range resource.Actions {
				for _, relation := range action {
					relationSet[fmt.Sprintf("%s_%s", namespaceID, relation)] = true
				}
			}
			break
		}
	}

	if len(relationSet) == 0 {
		return nil, fmt.Errorf("resource not found")
	}

	return relationSet, err
}

func (repo *ResourcesRepository) refresh(ctx context.Context) error {
	var resources []resource.YAML

	// get all items
	it := repo.bucket.List(&blob.ListOptions{})
	for {
		obj, err := it.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if obj.IsDir {
			continue
		}
		if !(strings.HasSuffix(obj.Key, ".yaml") || strings.HasSuffix(obj.Key, ".yml")) {
			continue
		}
		fileBytes, err := repo.bucket.ReadAll(ctx, obj.Key)
		if err != nil {
			return errors.Wrap(err, "bucket.ReadAll: "+obj.Key)
		}

		var resourceBackends ResourceBackends
		if err := yaml.Unmarshal(fileBytes, &resourceBackends); err != nil {
			return errors.Wrap(err, "yaml.Unmarshal: "+obj.Key)
		}
		if len(resourceBackends.Backends) == 0 {
			continue
		}

		for _, resourceBackend := range resourceBackends.Backends {
			for _, resourceType := range resourceBackend.ResourceTypes {
				resources = append(resources, resource.YAML{
					Name:    namespace.CreateID(resourceBackend.Name, resourceType.Name),
					Actions: resourceType.Actions,
				})
			}
		}
	}

	repo.mu.Lock()
	repo.cached = resources
	repo.mu.Unlock()
	repo.log.Debug("resource config cache refreshed", "resource_config_count", len(repo.cached))
	return nil
}

func (repo *ResourcesRepository) InitCache(ctx context.Context, refreshDelay time.Duration) error {
	repo.cron = cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
	))
	if _, err := repo.cron.AddFunc("@every "+refreshDelay.String(), func() {
		if err := repo.refresh(ctx); err != nil {
			repo.log.Warn("failed to refresh resource config repository", "err", err)
		}
	}); err != nil {
		return err
	}
	repo.cron.Start()

	// do it once right now
	return repo.refresh(ctx)
}

func (repo *ResourcesRepository) Close() error {
	<-repo.cron.Stop().Done()
	return repo.bucket.Close()
}

func NewResourcesRepository(logger log.Logger, b Bucket) *ResourcesRepository {
	return &ResourcesRepository{
		log:    logger,
		bucket: b,
		mu:     new(sync.Mutex),
	}
}
