package blob

import (
	"context"
	"io"
	"regexp"
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

type RuleRepository struct {
	log log.Logger
	mu  *sync.Mutex

	cron   *cron.Cron
	bucket store.Bucket
	cached []structs.Ruleset
}

func (repo *RuleRepository) GetAll(ctx context.Context) ([]structs.Ruleset, error) {
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

func (repo *RuleRepository) refresh(ctx context.Context) error {
	var ruleset []structs.Ruleset

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

		var s structs.Ruleset
		if err := yaml.Unmarshal(fileBytes, &s); err != nil {
			return errors.Wrap(err, "yaml.Unmarshal: "+obj.Key)
		}
		if len(s.Rules) == 0 {
			continue
		}

		// parse all urls at this time only to avoid doing it usage
		var rxParsingSuccess = true
		for ruleIdx, rule := range s.Rules {
			// TODO: only compile between delimiter, maybe angular brackets
			s.Rules[ruleIdx].Frontend.URLRx, err = regexp.Compile(rule.Frontend.URL)
			if err != nil {
				rxParsingSuccess = false
				repo.log.Error("failed to parse rule frontend as a valid regular expression",
					"url", rule.Frontend.URL, "err", err)
			}
		}

		if rxParsingSuccess {
			ruleset = append(ruleset, s)
		} else {
			repo.log.Warn("skipping rule set due to parsing errors", "content", string(fileBytes))
		}
	}

	repo.mu.Lock()
	repo.cached = ruleset
	repo.mu.Unlock()
	repo.log.Debug("rule cache refreshed", "ruleset_count", len(repo.cached))
	return nil
}

func (repo *RuleRepository) InitCache(ctx context.Context, refreshDelay time.Duration) error {
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

func (repo *RuleRepository) Close() error {
	<-repo.cron.Stop().Done()
	return repo.bucket.Close()
}

func NewRuleRepository(logger log.Logger, b store.Bucket) *RuleRepository {
	return &RuleRepository{
		log:    logger,
		bucket: b,
		mu:     new(sync.Mutex),
	}
}
