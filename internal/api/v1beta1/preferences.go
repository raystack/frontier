package v1beta1

import (
	"context"

	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PreferenceService interface {
	Create(ctx context.Context, preference preference.Preference) (preference.Preference, error)
	Describe(ctx context.Context) []preference.Trait
	List(ctx context.Context, filter preference.Filter) ([]preference.Preference, error)
	LoadPlatformPreferences(ctx context.Context) (map[string]string, error)
}

func (h Handler) ListPreferences(ctx context.Context, in *frontierv1beta1.ListPreferencesRequest) (*frontierv1beta1.ListPreferencesResponse, error) {
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		ResourceID:   preference.PlatformID,
		ResourceType: schema.PlatformNamespace,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range prefs {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}
	return &frontierv1beta1.ListPreferencesResponse{
		Preferences: pbPrefs,
	}, nil
}

func (h Handler) CreatePreferences(ctx context.Context, request *frontierv1beta1.CreatePreferencesRequest) (*frontierv1beta1.CreatePreferencesResponse, error) {
	var createdPreferences []preference.Preference
	for _, prefBody := range request.GetPreferences() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.GetName(),
			Value:        prefBody.GetValue(),
			ResourceID:   preference.PlatformID,
			ResourceType: schema.PlatformNamespace,
		})
		if err != nil {
			if errors.Is(err, preference.ErrTraitNotFound) {
				return nil, status.Errorf(codes.InvalidArgument, err.Error())
			}
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		createdPreferences = append(createdPreferences, pref)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range createdPreferences {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}

	return &frontierv1beta1.CreatePreferencesResponse{
		Preference: pbPrefs,
	}, nil
}

func (h Handler) DescribePreferences(ctx context.Context, request *frontierv1beta1.DescribePreferencesRequest) (*frontierv1beta1.DescribePreferencesResponse, error) {
	prefTraits := h.preferenceService.Describe(ctx)
	var pbTraits []*frontierv1beta1.PreferenceTrait
	for _, trait := range prefTraits {
		pbTraits = append(pbTraits, transformPreferenceTraitToPB(trait))
	}
	return &frontierv1beta1.DescribePreferencesResponse{
		Traits: pbTraits,
	}, nil
}

func (h Handler) CreateOrganizationPreferences(ctx context.Context, request *frontierv1beta1.CreateOrganizationPreferencesRequest) (*frontierv1beta1.CreateOrganizationPreferencesResponse, error) {
	var createdPreferences []preference.Preference
	for _, prefBody := range request.GetBodies() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.GetName(),
			Value:        prefBody.GetValue(),
			ResourceID:   request.GetId(),
			ResourceType: schema.OrganizationNamespace,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		createdPreferences = append(createdPreferences, pref)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range createdPreferences {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}
	return &frontierv1beta1.CreateOrganizationPreferencesResponse{
		Preferences: pbPrefs,
	}, nil
}

func (h Handler) ListOrganizationPreferences(ctx context.Context, request *frontierv1beta1.ListOrganizationPreferencesRequest) (*frontierv1beta1.ListOrganizationPreferencesResponse, error) {
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		OrgID: request.GetId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range prefs {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}
	return &frontierv1beta1.ListOrganizationPreferencesResponse{
		Preferences: pbPrefs,
	}, nil
}

func (h Handler) CreateProjectPreferences(ctx context.Context, request *frontierv1beta1.CreateProjectPreferencesRequest) (*frontierv1beta1.CreateProjectPreferencesResponse, error) {
	// TODO implement me
	return nil, grpcOperationUnsupported
}

func (h Handler) ListProjectPreferences(ctx context.Context, request *frontierv1beta1.ListProjectPreferencesRequest) (*frontierv1beta1.ListProjectPreferencesResponse, error) {
	// TODO implement me
	return nil, grpcOperationUnsupported
}

func (h Handler) CreateGroupPreferences(ctx context.Context, request *frontierv1beta1.CreateGroupPreferencesRequest) (*frontierv1beta1.CreateGroupPreferencesResponse, error) {
	// TODO implement me
	return nil, grpcOperationUnsupported
}

func (h Handler) ListGroupPreferences(ctx context.Context, request *frontierv1beta1.ListGroupPreferencesRequest) (*frontierv1beta1.ListGroupPreferencesResponse, error) {
	// TODO implement me
	return nil, grpcOperationUnsupported
}

func (h Handler) CreateUserPreferences(ctx context.Context, request *frontierv1beta1.CreateUserPreferencesRequest) (*frontierv1beta1.CreateUserPreferencesResponse, error) {
	var createdPreferences []preference.Preference
	for _, prefBody := range request.GetBodies() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.GetName(),
			Value:        prefBody.GetValue(),
			ResourceID:   request.GetId(),
			ResourceType: schema.UserPrincipal,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		createdPreferences = append(createdPreferences, pref)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range createdPreferences {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}
	return &frontierv1beta1.CreateUserPreferencesResponse{
		Preferences: pbPrefs,
	}, nil
}

func (h Handler) ListUserPreferences(ctx context.Context, request *frontierv1beta1.ListUserPreferencesRequest) (*frontierv1beta1.ListUserPreferencesResponse, error) {
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		UserID: request.GetId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range prefs {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}
	return &frontierv1beta1.ListUserPreferencesResponse{
		Preferences: pbPrefs,
	}, nil
}

func (h Handler) CreateCurrentUserPreferences(ctx context.Context, request *frontierv1beta1.CreateCurrentUserPreferencesRequest) (*frontierv1beta1.CreateCurrentUserPreferencesResponse, error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}

	var createdPreferences []preference.Preference
	for _, prefBody := range request.GetBodies() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.GetName(),
			Value:        prefBody.GetValue(),
			ResourceID:   principal.ID,
			ResourceType: schema.UserPrincipal,
		})
		if err != nil {
			switch {
			case errors.Is(err, preference.ErrTraitNotFound), errors.Is(err, preference.ErrInvalidValue):
				return nil, status.Errorf(codes.InvalidArgument, err.Error())
			default:
				return nil, status.Errorf(codes.Internal, err.Error())
			}
		}
		createdPreferences = append(createdPreferences, pref)
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range createdPreferences {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}
	return &frontierv1beta1.CreateCurrentUserPreferencesResponse{
		Preferences: pbPrefs,
	}, nil
}

func (h Handler) ListCurrentUserPreferences(ctx context.Context, request *frontierv1beta1.ListCurrentUserPreferencesRequest) (*frontierv1beta1.ListCurrentUserPreferencesResponse, error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		UserID: principal.ID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var pbPrefs []*frontierv1beta1.Preference
	for _, pref := range prefs {
		pbPrefs = append(pbPrefs, transformPreferenceToPB(pref))
	}
	return &frontierv1beta1.ListCurrentUserPreferencesResponse{
		Preferences: pbPrefs,
	}, nil
}

func (h Handler) ListPlatformPreferences(ctx context.Context) (map[string]string, error) {
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
