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
		return nil, handlePreferenceError(err)
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
			return nil, handlePreferenceError(err)
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

func (h *ConnectHandler) CreateOrganizationPreferences(ctx context.Context, req *connect.Request[frontierv1beta1.CreateOrganizationPreferencesRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationPreferencesResponse], error) {
	var createdPreferences []preference.Preference
	for _, prefBody := range req.Msg.GetBodies() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.GetName(),
			Value:        prefBody.GetValue(),
			ResourceID:   req.Msg.GetId(),
			ResourceType: schema.OrganizationNamespace,
		})
		if err != nil {
			return nil, handlePreferenceError(err)
		}
		createdPreferences = append(createdPreferences, pref)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range createdPreferences {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}

	return connect.NewResponse(&frontierv1beta1.CreateOrganizationPreferencesResponse{
		Preferences: pbPrefs,
	}), nil
}

func (h *ConnectHandler) ListOrganizationPreferences(ctx context.Context, req *connect.Request[frontierv1beta1.ListOrganizationPreferencesRequest]) (*connect.Response[frontierv1beta1.ListOrganizationPreferencesResponse], error) {
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		OrgID: req.Msg.GetId(),
	})
	if err != nil {
		return nil, handlePreferenceError(err)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range prefs {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationPreferencesResponse{
		Preferences: pbPrefs,
	}), nil
}

func (h *ConnectHandler) CreateUserPreferences(ctx context.Context, req *connect.Request[frontierv1beta1.CreateUserPreferencesRequest]) (*connect.Response[frontierv1beta1.CreateUserPreferencesResponse], error) {
	var createdPreferences []preference.Preference
	for _, prefBody := range req.Msg.GetBodies() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.GetName(),
			Value:        prefBody.GetValue(),
			ResourceID:   req.Msg.GetId(),
			ResourceType: schema.UserPrincipal,
		})
		if err != nil {
			return nil, handlePreferenceError(err)
		}
		createdPreferences = append(createdPreferences, pref)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range createdPreferences {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}

	return connect.NewResponse(&frontierv1beta1.CreateUserPreferencesResponse{
		Preferences: pbPrefs,
	}), nil
}

func (h *ConnectHandler) ListUserPreferences(ctx context.Context, req *connect.Request[frontierv1beta1.ListUserPreferencesRequest]) (*connect.Response[frontierv1beta1.ListUserPreferencesResponse], error) {
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		UserID: req.Msg.GetId(),
	})
	if err != nil {
		return nil, handlePreferenceError(err)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range prefs {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}

	return connect.NewResponse(&frontierv1beta1.ListUserPreferencesResponse{
		Preferences: pbPrefs,
	}), nil
}

func (h *ConnectHandler) CreateCurrentUserPreferences(ctx context.Context, req *connect.Request[frontierv1beta1.CreateCurrentUserPreferencesRequest]) (*connect.Response[frontierv1beta1.CreateCurrentUserPreferencesResponse], error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}

	var createdPreferences []preference.Preference
	for _, prefBody := range req.Msg.GetBodies() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.GetName(),
			Value:        prefBody.GetValue(),
			ResourceID:   principal.ID,
			ResourceType: schema.UserPrincipal,
		})
		if err != nil {
			return nil, handlePreferenceError(err)
		}
		createdPreferences = append(createdPreferences, pref)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range createdPreferences {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}

	return connect.NewResponse(&frontierv1beta1.CreateCurrentUserPreferencesResponse{
		Preferences: pbPrefs,
	}), nil
}

func (h *ConnectHandler) ListCurrentUserPreferences(ctx context.Context, req *connect.Request[frontierv1beta1.ListCurrentUserPreferencesRequest]) (*connect.Response[frontierv1beta1.ListCurrentUserPreferencesResponse], error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}

	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		UserID: principal.ID,
	})
	if err != nil {
		return nil, handlePreferenceError(err)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range prefs {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}

	return connect.NewResponse(&frontierv1beta1.ListCurrentUserPreferencesResponse{
		Preferences: pbPrefs,
	}), nil
}

func (h *ConnectHandler) DescribePreferences(ctx context.Context, req *connect.Request[frontierv1beta1.DescribePreferencesRequest]) (*connect.Response[frontierv1beta1.DescribePreferencesResponse], error) {
	prefTraits := h.preferenceService.Describe(ctx)
	var pbTraits []*frontierv1beta1.PreferenceTrait
	for _, trait := range prefTraits {
		pbTraits = append(pbTraits, transformPreferenceTraitToPB(trait))
	}
	return connect.NewResponse(&frontierv1beta1.DescribePreferencesResponse{
		Traits: pbTraits,
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

func transformPreferenceTraitToPB(pref preference.Trait) *frontierv1beta1.PreferenceTrait {
	pbTrait := &frontierv1beta1.PreferenceTrait{
		ResourceType:    pref.ResourceType,
		Name:            pref.Name,
		Title:           pref.Title,
		Description:     pref.Description,
		LongDescription: pref.LongDescription,
		Heading:         pref.Heading,
		SubHeading:      pref.SubHeading,
		Breadcrumb:      pref.Breadcrumb,
		InputHints:      pref.InputHints,
		Default:         pref.Default,
	}
	switch pref.Input {
	case preference.TraitInputText:
		pbTrait.InputType = frontierv1beta1.PreferenceTrait_INPUT_TYPE_TEXT
	case preference.TraitInputTextarea:
		pbTrait.InputType = frontierv1beta1.PreferenceTrait_INPUT_TYPE_TEXTAREA
	case preference.TraitInputSelect:
		pbTrait.InputType = frontierv1beta1.PreferenceTrait_INPUT_TYPE_SELECT
	case preference.TraitInputCombobox:
		pbTrait.InputType = frontierv1beta1.PreferenceTrait_INPUT_TYPE_COMBOBOX
	case preference.TraitInputCheckbox:
		pbTrait.InputType = frontierv1beta1.PreferenceTrait_INPUT_TYPE_CHECKBOX
	case preference.TraitInputMultiselect:
		pbTrait.InputType = frontierv1beta1.PreferenceTrait_INPUT_TYPE_MULTISELECT
	case preference.TraitInputNumber:
		pbTrait.InputType = frontierv1beta1.PreferenceTrait_INPUT_TYPE_NUMBER
	}
	return pbTrait
}

func handlePreferenceError(err error) *connect.Error {
	switch {
	case errors.Is(err, preference.ErrInvalidFilter):
		return connect.NewError(connect.CodeInvalidArgument, ErrInvalidPreferenceFilter)
	case errors.Is(err, preference.ErrTraitNotFound):
		return connect.NewError(connect.CodeInvalidArgument, ErrTraitNotFound)
	case errors.Is(err, preference.ErrInvalidValue):
		return connect.NewError(connect.CodeInvalidArgument, ErrInvalidPreferenceValue)
	default:
		return connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
}
