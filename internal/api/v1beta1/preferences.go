package v1beta1

import (
	"context"

	"github.com/google/uuid"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PreferenceService interface {
	Create(ctx context.Context, preference preference.Preference) (preference.Preference, error)
	Describe(ctx context.Context) []preference.Trait
	List(ctx context.Context, filter preference.Filter) ([]preference.Preference, error)
}

func (h Handler) ListPreferences(ctx context.Context, in *frontierv1beta1.ListPreferencesRequest) (*frontierv1beta1.ListPreferencesResponse, error) {
	logger := grpczap.Extract(ctx)
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		ResourceID: uuid.Nil.String(), // nil UUID for a platform-wide preference
	})
	if err != nil {
		logger.Error(err.Error())
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
	logger := grpczap.Extract(ctx)
	var createdPreferences []preference.Preference
	for _, prefBody := range request.GetPreferences() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.Name,
			Value:        prefBody.Value,
			ResourceID:   uuid.Nil.String(), // nil UUID for a platform-wide preference
			ResourceType: schema.PlatformNamespace,
		})
		if err != nil {
			logger.Error(err.Error())
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
	logger := grpczap.Extract(ctx)
	var createdPreferences []preference.Preference
	for _, prefBody := range request.GetBodies() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.Name,
			Value:        prefBody.Value,
			ResourceID:   request.GetId(),
			ResourceType: schema.OrganizationNamespace,
		})
		if err != nil {
			logger.Error(err.Error())
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
	logger := grpczap.Extract(ctx)
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		OrgID: request.GetId(),
	})
	if err != nil {
		logger.Error(err.Error())
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
	//TODO implement me
	return nil, grpcOperationUnsupported
}

func (h Handler) ListProjectPreferences(ctx context.Context, request *frontierv1beta1.ListProjectPreferencesRequest) (*frontierv1beta1.ListProjectPreferencesResponse, error) {
	//TODO implement me
	return nil, grpcOperationUnsupported
}

func (h Handler) CreateGroupPreferences(ctx context.Context, request *frontierv1beta1.CreateGroupPreferencesRequest) (*frontierv1beta1.CreateGroupPreferencesResponse, error) {
	//TODO implement me
	return nil, grpcOperationUnsupported
}

func (h Handler) ListGroupPreferences(ctx context.Context, request *frontierv1beta1.ListGroupPreferencesRequest) (*frontierv1beta1.ListGroupPreferencesResponse, error) {
	//TODO implement me
	return nil, grpcOperationUnsupported
}

func (h Handler) CreateUserPreferences(ctx context.Context, request *frontierv1beta1.CreateUserPreferencesRequest) (*frontierv1beta1.CreateUserPreferencesResponse, error) {
	logger := grpczap.Extract(ctx)
	var createdPreferences []preference.Preference
	for _, prefBody := range request.GetBodies() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.Name,
			Value:        prefBody.Value,
			ResourceID:   request.GetId(),
			ResourceType: schema.UserPrincipal,
		})
		if err != nil {
			logger.Error(err.Error())
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
	logger := grpczap.Extract(ctx)
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		UserID: request.GetId(),
	})
	if err != nil {
		logger.Error(err.Error())
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
	logger := grpczap.Extract(ctx)
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}

	var createdPreferences []preference.Preference
	for _, prefBody := range request.GetBodies() {
		pref, err := h.preferenceService.Create(ctx, preference.Preference{
			Name:         prefBody.Name,
			Value:        prefBody.Value,
			ResourceID:   principal.ID,
			ResourceType: schema.UserPrincipal,
		})
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, err.Error())
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
	logger := grpczap.Extract(ctx)
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	prefs, err := h.preferenceService.List(ctx, preference.Filter{
		UserID: principal.ID,
	})
	if err != nil {
		logger.Error(err.Error())
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
