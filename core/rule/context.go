package rule

import (
	"context"
)

type contextRuleKey struct{}

func WithContext(ctx context.Context, rule *Rule) context.Context {
	return context.WithValue(ctx, contextRuleKey{}, rule)
}

func GetFromContext(ctx context.Context) (*Rule, bool) {
	rl, ok := ctx.Value(contextRuleKey{}).(*Rule)
	return rl, ok
}
