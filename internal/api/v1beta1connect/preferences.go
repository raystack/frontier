package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) ListPreferences(ctx context.Context, req *connect.Request[frontierv1beta1.ListPreferencesRequest]) (*connect.Response[frontierv1beta1.ListPreferencesResponse], error) {
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		ResourceID:   preference.PlatformID,
		ResourceType: schema.PlatformNamespace,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range prefs {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}
	return connect.NewResponse(&frontierv1beta1.ListPreferencesResponse{
		Preferences: pbPrefs,
	}), nil
}

func (h *ConnectHandler) CreatePreferences(ctx context.Context, req *connect.Request[frontierv1beta1.CreatePreferencesRequest]) (*connect.Response[frontierv1beta1.CreatePreferencesResponse], error) {
	var createdPreferences []preference.Preference
	for _, prefBody := range req.Msg.GetPreferences() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.GetName(),
			Value:        prefBody.GetValue(),
			ResourceID:   preference.PlatformID,
			ResourceType: schema.PlatformNamespace,
		})
		if err != nil {
			if errors.Is(err, preference.ErrTraitNotFound) {
				return nil, connect.NewError(connect.CodeInvalidArgument, err)
			}
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		createdPreferences = append(createdPreferences, pref)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range createdPreferences {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}

	return connect.NewResponse(&frontierv1beta1.CreatePreferencesResponse{
		Preference: pbPrefs,
	}), nil
}

func (h *ConnectHandler) ListPlatformPreferences(ctx context.Context) (map[string]string, error) {
	return h.preferenceService.LoadPlatformPreferences(ctx)
}

func transformPreferenceToPB(pref preference.Preference) *frontierv1beta1.Preference {
	return &frontierv1beta1.Preference{
		Id:           pref.ID,
		Name:         pref.Name,
		Value:        pref.Value,
		ResourceId:   pref.ResourceID,
		ResourceType: pref.ResourceType,
		CreatedAt:    timestamppb.New(pref.CreatedAt),
		UpdatedAt:    timestamppb.New(pref.UpdatedAt),
	}
}
