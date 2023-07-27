package cmd

import (
	"context"
	"errors"
	"net/http"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/rule"
	"github.com/raystack/frontier/internal/api/v1beta1"
	"github.com/raystack/frontier/internal/proxy"
	"github.com/raystack/frontier/internal/proxy/hook"
	authz_hook "github.com/raystack/frontier/internal/proxy/hook/authz"
	"github.com/raystack/frontier/internal/proxy/middleware/attributes"
	"github.com/raystack/frontier/internal/proxy/middleware/authz"
	"github.com/raystack/frontier/internal/proxy/middleware/basic_auth"
	"github.com/raystack/frontier/internal/proxy/middleware/observability"
	"github.com/raystack/frontier/internal/proxy/middleware/prefix"
	"github.com/raystack/frontier/internal/proxy/middleware/rulematch"
	"github.com/raystack/frontier/internal/store/blob"
	"github.com/raystack/salt/log"
)

func serveProxies(
	ctx context.Context,
	logger *log.Zap,
	identityProxyHeaderKey,
	userIDHeaderKey string,
	cfg proxy.ServicesConfig,
	resourceService *resource.Service,
	relationService *relation.Service,
	principalService *authenticate.Service,
	projectService *project.Service,
) ([]func() error, []func(ctx context.Context) error, error) {
	var cleanUpBlobs []func() error
	var cleanUpProxies []func(ctx context.Context) error

	for _, svcConfig := range cfg.Services {
		hookPipeline := buildHookPipeline(logger, resourceService, relationService, identityProxyHeaderKey)

		h2cProxy := proxy.NewH2c(
			proxy.NewH2cRoundTripper(logger, hookPipeline),
			proxy.NewDirector(),
		)

		// load rules sets
		if svcConfig.RulesPath == "" {
			return nil, nil, errors.New("ruleset field cannot be left empty")
		}

		ruleBlobFS, err := blob.NewStore(ctx, svcConfig.RulesPath, svcConfig.RulesPathSecret)
		if err != nil {
			return nil, nil, err
		}

		ruleBlobRepository := blob.NewRuleRepository(logger, ruleBlobFS)
		if err := ruleBlobRepository.InitCache(ctx, ruleCacheRefreshDelay); err != nil {
			return nil, nil, err
		}
		cleanUpBlobs = append(cleanUpBlobs, ruleBlobRepository.Close)

		ruleService := rule.NewService(ruleBlobRepository)

		middlewarePipeline := buildMiddlewarePipeline(logger, h2cProxy, identityProxyHeaderKey, userIDHeaderKey,
			resourceService, principalService, ruleService, projectService)

		cps := proxy.Serve(ctx, logger, svcConfig, middlewarePipeline)
		cleanUpProxies = append(cleanUpProxies, cps)
		logger.Info("[frontier] proxy is up")
	}
	return cleanUpBlobs, cleanUpProxies, nil
}

func buildHookPipeline(log log.Logger, resourceService v1beta1.ResourceService, relationService v1beta1.RelationService, identityProxyHeaderKey string) hook.Service {
	rootHook := hook.New()
	return authz_hook.New(log, rootHook, rootHook, resourceService, relationService, identityProxyHeaderKey)
}

// buildPipeline builds middleware sequence
func buildMiddlewarePipeline(
	logger *log.Zap,
	proxy http.Handler,
	identityProxyHeaderKey, userIDHeaderKey string,
	resourceService *resource.Service,
	principalService *authenticate.Service,
	ruleService *rule.Service,
	projectService *project.Service,
) http.Handler {
	// Note: execution order is bottom up
	prefixWare := prefix.New(logger, proxy)
	spiceDBAuthz := authz.New(logger, prefixWare, userIDHeaderKey, resourceService, principalService)
	basicAuthn := basic_auth.New(logger, spiceDBAuthz)
	attributeExtractor := attributes.New(logger, basicAuthn, identityProxyHeaderKey, projectService)
	matchWare := rulematch.New(logger, attributeExtractor, rulematch.NewRouteMatcher(ruleService))
	observability := observability.New(logger, matchWare)
	return observability
}
