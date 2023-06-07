package interceptors

import (
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// GatewayHeaderMatcherFunc allows bypassing default runtime behaviour of prefixing headers with `grpc-gateway`
func GatewayHeaderMatcherFunc(headerKeys map[string]bool) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		if _, ok := headerKeys[strings.ToLower(key)]; ok {
			return key, true
		}
		return runtime.DefaultHeaderMatcher(key)
	}
}
