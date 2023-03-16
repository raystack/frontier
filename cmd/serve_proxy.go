package cmd

import (
	"context"
	"errors"
	"net/http"

	"github.com/goto/salt/log"
	"github.com/goto/shield/core/project"
	"github.com/goto/shield/core/relation"
	"github.com/goto/shield/core/resource"
	"github.com/goto/shield/core/rule"
	"github.com/goto/shield/core/user"
	"github.com/goto/shield/internal/api/v1beta1"
	"github.com/goto/shield/internal/proxy"
	"github.com/goto/shield/internal/proxy/hook"
	authz_hook "github.com/goto/shield/internal/proxy/hook/authz"
	"github.com/goto/shield/internal/proxy/middleware/attributes"
	"github.com/goto/shield/internal/proxy/middleware/authz"
	"github.com/goto/shield/internal/proxy/middleware/basic_auth"
	"github.com/goto/shield/internal/proxy/middleware/observability"
	"github.com/goto/shield/internal/proxy/middleware/prefix"
	"github.com/goto/shield/internal/proxy/middleware/rulematch"
	"github.com/goto/shield/internal/store/blob"
)

func serveProxies(
	ctx context.Context,
	logger *log.Zap,
	identityProxyHeaderKey,
	userIDHeaderKey string,
	cfg proxy.ServicesConfig,
	resourceService *resource.Service,
	relationService *relation.Service,
	userService *user.Service,
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

		middlewarePipeline := buildMiddlewarePipeline(logger, h2cProxy, identityProxyHeaderKey, userIDHeaderKey, resourceService, userService, ruleService, projectService)

		cps := proxy.Serve(ctx, logger, svcConfig, middlewarePipeline)
		cleanUpProxies = append(cleanUpProxies, cps)
	}

	logger.Info("[shield] proxy is up")
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
	userService *user.Service,
	ruleService *rule.Service,
	projectService *project.Service,
) http.Handler {
	// Note: execution order is bottom up
	prefixWare := prefix.New(logger, proxy)
	casbinAuthz := authz.New(logger, prefixWare, userIDHeaderKey, resourceService, userService)
	basicAuthn := basic_auth.New(logger, casbinAuthz)
	attributeExtractor := attributes.New(logger, basicAuthn, identityProxyHeaderKey, projectService)
	matchWare := rulematch.New(logger, attributeExtractor, rulematch.NewRouteMatcher(ruleService))
	observability := observability.New(logger, matchWare)
	return observability
}
