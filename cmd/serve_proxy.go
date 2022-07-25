package cmd

import (
	"context"
	"errors"
	"net/http"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/rule"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/api/v1beta1"
	"github.com/odpf/shield/internal/proxy"
	"github.com/odpf/shield/internal/proxy/hook"
	authz_hook "github.com/odpf/shield/internal/proxy/hook/authz"
	"github.com/odpf/shield/internal/proxy/middleware/authz"
	"github.com/odpf/shield/internal/proxy/middleware/basic_auth"
	"github.com/odpf/shield/internal/proxy/middleware/prefix"
	"github.com/odpf/shield/internal/proxy/middleware/rulematch"
	"github.com/odpf/shield/internal/store/blob"
)

func serveProxies(
	ctx context.Context,
	logger log.Logger,
	identityProxyHeader string,
	cfg proxy.ServicesConfig,
	resourceService *resource.Service,
	userService *user.Service,
) ([]func() error, []func(ctx context.Context) error, error) {
	var cleanUpBlobs []func() error
	var cleanUpProxies []func(ctx context.Context) error

	for _, svcConfig := range cfg.Services {
		hookPipeline := buildHookPipeline(logger, identityProxyHeader, resourceService)

		h2cProxy := proxy.NewH2c(
			proxy.NewH2cRoundTripper(logger, hookPipeline),
			proxy.NewDirector(),
		)

		// load rules sets
		if svcConfig.RulesPath == "" {
			return nil, nil, errors.New("ruleset field cannot be left empty")
		}

		blobFS, err := blob.NewStore(ctx, svcConfig.RulesPath, svcConfig.RulesPathSecret)
		if err != nil {
			return nil, nil, err
		}

		ruleRepo := blob.NewRuleRepository(logger, blobFS)
		if err := ruleRepo.InitCache(ctx, ruleCacheRefreshDelay); err != nil {
			return nil, nil, err
		}

		ruleService := rule.NewService(ruleRepo)

		cleanUpBlobs = append(cleanUpBlobs, ruleRepo.Close)

		middlewarePipeline := buildMiddlewarePipeline(logger, h2cProxy, identityProxyHeader, resourceService, userService, ruleService)

		cps := proxy.Serve(ctx, logger, svcConfig, middlewarePipeline)
		cleanUpProxies = append(cleanUpProxies, cps)
	}

	logger.Info("[shield] proxy is up")
	return cleanUpBlobs, cleanUpProxies, nil
}

func buildHookPipeline(log log.Logger, identityProxyHeader string, resourceService v1beta1.ResourceService) hook.Service {
	rootHook := hook.New()
	return authz_hook.New(log, rootHook, rootHook, identityProxyHeader, resourceService)
}

// buildPipeline builds middleware sequence
func buildMiddlewarePipeline(
	logger log.Logger,
	proxy http.Handler,
	identityProxyHeader string,
	resourceService *resource.Service,
	userService *user.Service,
	ruleService *rule.Service,
) http.Handler {
	// Note: execution order is bottom up
	prefixWare := prefix.New(logger, proxy)
	casbinAuthz := authz.New(logger, prefixWare, identityProxyHeader, resourceService, userService)
	basicAuthn := basic_auth.New(logger, casbinAuthz)
	matchWare := rulematch.New(logger, basicAuthn, rulematch.NewRouteMatcher(ruleService))
	return matchWare
}
