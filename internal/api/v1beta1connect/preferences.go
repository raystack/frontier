package v1beta1connect

import (
	"context"
)

func (h *ConnectHandler) ListPlatformPreferences(ctx context.Context) (map[string]string, error) {
	return h.preferenceService.LoadPlatformPreferences(ctx)
}
