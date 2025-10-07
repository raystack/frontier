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

	return connect.NewResponse(&frontierv1beta1.CreateOrganizationPreferencesResponse{
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
		pbTrait.Input = &frontierv1beta1.PreferenceTrait_Text{}
	case preference.TraitInputTextarea:
		pbTrait.Input = &frontierv1beta1.PreferenceTrait_Textarea{}
	case preference.TraitInputSelect:
		pbTrait.Input = &frontierv1beta1.PreferenceTrait_Select{}
	case preference.TraitInputCombobox:
		pbTrait.Input = &frontierv1beta1.PreferenceTrait_Combobox{}
	case preference.TraitInputCheckbox:
		pbTrait.Input = &frontierv1beta1.PreferenceTrait_Checkbox{}
	case preference.TraitInputMultiselect:
		pbTrait.Input = &frontierv1beta1.PreferenceTrait_Multiselect{}
	case preference.TraitInputNumber:
		pbTrait.Input = &frontierv1beta1.PreferenceTrait_Number{}
	}
	return pbTrait
}
