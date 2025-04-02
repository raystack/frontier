package blob

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/goto/salt/log"

	"github.com/ghodss/yaml"
	"github.com/odpf/shield/store"
	"github.com/odpf/shield/structs"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type Ruleset struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Backends []Backend `yaml:"backends"`
}

type Backend struct {
	Name      string     `yaml:"name"`
	Target    string     `yaml:"target"`
	Methods   []string   `yaml:"methods"`
	Frontends []Frontend `yaml:"frontends"`
	Prefix    string     `yaml:"prefix"`
}

type Frontend struct {
	Action      string       `yaml:"action"`
	Path        string       `yaml:"path"`
	Method      string       `yaml:"method"`
	Middlewares []Middleware `yaml:"middlewares"`
	Hooks       []Hook       `yaml:"hooks"`
}

type Middleware struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:"config"`
}

type Hook struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:"config"`
}

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

	files, err := repo.bucket.ListFiles(ctx, "")
	if err != nil {
		return err
	}

	for _, fileKey := range files {
		if !(strings.HasSuffix(fileKey, ".yaml") || strings.HasSuffix(fileKey, ".yml")) {
			continue
		}

		fileBytes, err := repo.bucket.ReadFile(ctx, fileKey) // Fix function name
		if err != nil {
			return errors.Wrap(err, "bucket.ReadFile: "+fileKey) // Fix error message
		}

		var s Ruleset
		if err := yaml.Unmarshal(fileBytes, &s); err != nil {
			return errors.Wrap(err, "yaml.Unmarshal: "+fileKey)
		}
		if len(s.Rules) == 0 {
			continue
		}

		// transforming yaml parse ruleset to clean iterable ruleset in middlewares
		targetRuleSet := structs.Ruleset{}
		for _, rule := range s.Rules {
			for _, backend := range rule.Backends {
				for _, frontend := range backend.Frontends {
					middlewares := structs.MiddlewareSpecs{}
					for _, middleware := range frontend.Middlewares {
						middlewares = append(middlewares, structs.MiddlewareSpec{
							Name:   middleware.Name,
							Config: middleware.Config,
						})
					}

					hooks := structs.HookSpecs{}
					for _, hook := range frontend.Hooks {
						hooks = append(hooks, structs.HookSpec{
							Name:   hook.Name,
							Config: hook.Config,
						})
					}

					targetRuleSet.Rules = append(targetRuleSet.Rules, structs.Rule{
						Frontend: structs.Frontend{
							URL:    frontend.Path,
							Method: frontend.Method,
						},
						Backend:     structs.Backend{URL: backend.Target, Namespace: backend.Name, Prefix: backend.Prefix},
						Middlewares: middlewares,
						Hooks:       hooks,
					})
				}
			}
		}

		// parse all urls at this time only to avoid doing it usage
		var rxParsingSuccess = true
		for ruleIdx, rule := range targetRuleSet.Rules {
			// TODO: only compile between delimiter, maybe angular brackets
			targetRuleSet.Rules[ruleIdx].Frontend.URLRx, err = regexp.Compile(rule.Frontend.URL)
			if err != nil {
				rxParsingSuccess = false
				repo.log.Error("failed to parse rule frontend as a valid regular expression",
					"url", rule.Frontend.URL, "err", err)
			}
		}

		if rxParsingSuccess {
			ruleset = append(ruleset, targetRuleSet)
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
