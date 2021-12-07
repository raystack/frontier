package blob

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/odpf/salt/log"

	"github.com/robfig/cron/v3"

	"github.com/ghodss/yaml"
	"github.com/odpf/shield/store"
	"github.com/odpf/shield/structs"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
)

type ResourcesRepository struct {
	log log.Logger
	mu  *sync.Mutex

	cron   *cron.Cron
	bucket store.Bucket
	cached []structs.Resources
}

func (repo *ResourcesRepository) GetAll(ctx context.Context) ([]structs.Resources, error) {
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

func (repo *ResourcesRepository) refresh(ctx context.Context) error {
	var resources []structs.Resources

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

		var resource structs.Resources
		if err := yaml.Unmarshal(fileBytes, &resource); err != nil {
			return errors.Wrap(err, "yaml.Unmarshal: "+obj.Key)
		}
		if len(resource) == 0 {
			continue
		}

		resources = append(resources, resource)
	}

	repo.mu.Lock()
	repo.cached = resources
	repo.mu.Unlock()
	repo.log.Debug("rule cache refreshed", "ruleset_count", len(repo.cached))
	return nil
}

func (repo *ResourcesRepository) InitCache(ctx context.Context, refreshDelay time.Duration) error {
	repo.cron = cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
	))
	if _, err := repo.cron.AddFunc("@every "+refreshDelay.String(), func() {
		if err := repo.refresh(ctx); err != nil {
			repo.log.Warn("failed to refresh rule repository", "err", err)
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

func NewResourcesRepository(logger log.Logger, b store.Bucket) *ResourcesRepository {
	return &ResourcesRepository{
		log:    logger,
		bucket: b,
		mu:     new(sync.Mutex),
	}
}