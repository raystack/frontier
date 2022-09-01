package attributes

import "context"

type contextAttributesMap struct{}

func SetContextWithAttributes(ctx context.Context, requestAttributes map[string]any) context.Context {
	return context.WithValue(ctx, contextAttributesMap{}, requestAttributes)
}

func GetAttributesFromContext(ctx context.Context) (map[string]any, bool) {
	requestAttributes, ok := ctx.Value(contextAttributesMap{}).(map[string]any)
	return requestAttributes, ok
}
