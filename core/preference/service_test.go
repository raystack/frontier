package preference

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Set(ctx context.Context, preference Preference) (Preference, error) {
	args := m.Called(ctx, preference)
	return args.Get(0).(Preference), args.Error(1)
}

func (m *MockRepository) Get(ctx context.Context, id uuid.UUID) (Preference, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Preference), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, filter Filter) ([]Preference, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Preference), args.Error(1)
}

func TestLoadUserPreferences(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	orgID := "org-456"

	// Define test traits with defaults
	testTraits := []Trait{
		{
			ResourceType: schema.UserPrincipal,
			Name:         "theme",
			Default:      "light",
		},
		{
			ResourceType: schema.UserPrincipal,
			Name:         "language",
			Default:      "en",
		},
		{
			ResourceType: schema.UserPrincipal,
			Name:         "notifications",
			Default:      "", // No default
		},
		{
			ResourceType: schema.OrganizationNamespace, // Should be ignored for user prefs
			Name:         "org_setting",
			Default:      "default",
		},
	}

	// Define test traits with InputOptions for ValueDescription testing
	testTraitsWithOptions := []Trait{
		{
			ResourceType: schema.UserPrincipal,
			Name:         "unit_area",
			Default:      "sq_km",
			InputOptions: []InputHintOption{
				{Name: "sq_km", Description: "Square Kilometers"},
				{Name: "sq_ft", Description: "Square Feet"},
				{Name: "acres", Description: "Acres"},
			},
		},
		{
			ResourceType: schema.UserPrincipal,
			Name:         "theme",
			Default:      "light",
			// No InputOptions - ValueDescription should be empty
		},
	}

	t.Run("without scope returns global DB preferences plus defaults", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo, testTraits)

		filter := Filter{UserID: userID}
		dbPrefs := []Preference{
			{ID: "1", Name: "theme", Value: "dark", ResourceType: schema.UserPrincipal, ResourceID: userID},
		}
		mockRepo.On("List", ctx, filter).Return(dbPrefs, nil)

		result, err := svc.LoadUserPreferences(ctx, filter)

		assert.NoError(t, err)
		// Should have "theme" (from DB) and "language" (default)
		assert.Len(t, result, 2)

		resultMap := make(map[string]Preference)
		for _, p := range result {
			resultMap[p.Name] = p
		}

		assert.Equal(t, "dark", resultMap["theme"].Value)
		assert.Equal(t, "1", resultMap["theme"].ID)
		assert.Equal(t, "en", resultMap["language"].Value) // default
		assert.Equal(t, "", resultMap["language"].ID)      // no ID for default
		mockRepo.AssertExpectations(t)
	})

	t.Run("with scope returns complete preference set with scoped, global, and defaults", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo, testTraits)

		scopedFilter := Filter{
			UserID:    userID,
			ScopeType: schema.OrganizationNamespace,
			ScopeID:   orgID,
		}
		globalFilter := Filter{
			UserID: userID,
		}

		// Scoped: "theme" is set for this org
		scopedPrefs := []Preference{
			{ID: "1", Name: "theme", Value: "dark", ResourceType: schema.UserPrincipal, ResourceID: userID, ScopeType: schema.OrganizationNamespace, ScopeID: orgID},
		}
		// Global: "language" is set globally
		globalPrefs := []Preference{
			{ID: "2", Name: "language", Value: "fr", ResourceType: schema.UserPrincipal, ResourceID: userID},
		}

		mockRepo.On("List", ctx, scopedFilter).Return(scopedPrefs, nil)
		mockRepo.On("List", ctx, globalFilter).Return(globalPrefs, nil)

		result, err := svc.LoadUserPreferences(ctx, scopedFilter)

		assert.NoError(t, err)
		// Should have "theme" (scoped DB), "language" (global DB)
		// "notifications" has no default so should not be included
		assert.Len(t, result, 2)

		// Build a map for easier assertions
		resultMap := make(map[string]Preference)
		for _, p := range result {
			resultMap[p.Name] = p
		}

		// theme should be from scoped DB (dark)
		assert.Equal(t, "dark", resultMap["theme"].Value)
		assert.Equal(t, "1", resultMap["theme"].ID)

		// language should be from global DB (fr)
		assert.Equal(t, "fr", resultMap["language"].Value)
		assert.Equal(t, "2", resultMap["language"].ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("scoped preference takes priority over global", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo, testTraits)

		scopedFilter := Filter{
			UserID:    userID,
			ScopeType: schema.OrganizationNamespace,
			ScopeID:   orgID,
		}
		globalFilter := Filter{
			UserID: userID,
		}

		// Both scoped and global have "theme"
		scopedPrefs := []Preference{
			{ID: "1", Name: "theme", Value: "dark", ResourceType: schema.UserPrincipal, ResourceID: userID, ScopeType: schema.OrganizationNamespace, ScopeID: orgID},
		}
		globalPrefs := []Preference{
			{ID: "2", Name: "theme", Value: "light", ResourceType: schema.UserPrincipal, ResourceID: userID},
		}

		mockRepo.On("List", ctx, scopedFilter).Return(scopedPrefs, nil)
		mockRepo.On("List", ctx, globalFilter).Return(globalPrefs, nil)

		result, err := svc.LoadUserPreferences(ctx, scopedFilter)

		assert.NoError(t, err)

		resultMap := make(map[string]Preference)
		for _, p := range result {
			resultMap[p.Name] = p
		}

		// Scoped value should win
		assert.Equal(t, "dark", resultMap["theme"].Value)
		assert.Equal(t, "1", resultMap["theme"].ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("with scope but all prefs set returns no defaults", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo, testTraits)

		scopedFilter := Filter{
			UserID:    userID,
			ScopeType: schema.OrganizationNamespace,
			ScopeID:   orgID,
		}
		globalFilter := Filter{
			UserID: userID,
		}

		// All traits with defaults are set in scoped DB
		scopedPrefs := []Preference{
			{ID: "1", Name: "theme", Value: "dark", ResourceType: schema.UserPrincipal, ResourceID: userID},
			{ID: "2", Name: "language", Value: "de", ResourceType: schema.UserPrincipal, ResourceID: userID},
		}
		mockRepo.On("List", ctx, scopedFilter).Return(scopedPrefs, nil)
		mockRepo.On("List", ctx, globalFilter).Return([]Preference{}, nil)

		result, err := svc.LoadUserPreferences(ctx, scopedFilter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)

		resultMap := make(map[string]Preference)
		for _, p := range result {
			resultMap[p.Name] = p
		}

		// Both should be from DB
		assert.Equal(t, "dark", resultMap["theme"].Value)
		assert.Equal(t, "de", resultMap["language"].Value)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo, testTraits)

		filter := Filter{UserID: userID}
		mockRepo.On("List", ctx, filter).Return(nil, ErrInvalidFilter)

		result, err := svc.LoadUserPreferences(ctx, filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrInvalidFilter, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ValueDescription is populated from trait InputOptions for DB preferences", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo, testTraitsWithOptions)

		filter := Filter{UserID: userID}
		dbPrefs := []Preference{
			{ID: "1", Name: "unit_area", Value: "sq_ft", ResourceType: schema.UserPrincipal, ResourceID: userID},
		}
		mockRepo.On("List", ctx, filter).Return(dbPrefs, nil)

		result, err := svc.LoadUserPreferences(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2) // unit_area from DB + theme default

		resultMap := make(map[string]Preference)
		for _, p := range result {
			resultMap[p.Name] = p
		}

		// unit_area should have ValueDescription populated from InputOptions
		assert.Equal(t, "sq_ft", resultMap["unit_area"].Value)
		assert.Equal(t, "Square Feet", resultMap["unit_area"].ValueDescription)

		// theme has no InputOptions, so ValueDescription should be empty
		assert.Equal(t, "light", resultMap["theme"].Value)
		assert.Equal(t, "", resultMap["theme"].ValueDescription)

		mockRepo.AssertExpectations(t)
	})

	t.Run("ValueDescription is populated for default preferences", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo, testTraitsWithOptions)

		filter := Filter{UserID: userID}
		// No DB preferences - should get defaults
		mockRepo.On("List", ctx, filter).Return([]Preference{}, nil)

		result, err := svc.LoadUserPreferences(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2) // unit_area default + theme default

		resultMap := make(map[string]Preference)
		for _, p := range result {
			resultMap[p.Name] = p
		}

		// unit_area default should have ValueDescription from InputOptions
		assert.Equal(t, "sq_km", resultMap["unit_area"].Value)
		assert.Equal(t, "Square Kilometers", resultMap["unit_area"].ValueDescription)

		mockRepo.AssertExpectations(t)
	})
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	testTraits := []Trait{
		{
			ResourceType: schema.UserPrincipal,
			Name:         "unit_area",
			Input:        TraitInputSelect,
			Default:      "sq_km",
			InputOptions: []InputHintOption{
				{Name: "sq_km", Description: "Square Kilometers"},
				{Name: "sq_ft", Description: "Square Feet"},
				{Name: "acres", Description: "Acres"},
			},
		},
	}

	t.Run("Create populates ValueDescription from InputOptions", func(t *testing.T) {
		mockRepo := new(MockRepository)
		svc := NewService(mockRepo, testTraits)

		pref := Preference{
			Name:         "unit_area",
			Value:        "sq_ft",
			ResourceID:   userID,
			ResourceType: schema.UserPrincipal,
		}

		// Mock repo returns the preference as-is (like a real DB would)
		mockRepo.On("Set", ctx, pref).Return(Preference{
			ID:           "pref-123",
			Name:         "unit_area",
			Value:        "sq_ft",
			ResourceID:   userID,
			ResourceType: schema.UserPrincipal,
		}, nil)

		result, err := svc.Create(ctx, pref)

		require.NoError(t, err)
		assert.Equal(t, "sq_ft", result.Value)
		assert.Equal(t, "Square Feet", result.ValueDescription) // Should be populated from trait
		mockRepo.AssertExpectations(t)
	})
}
