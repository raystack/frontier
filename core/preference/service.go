package preference

import (
	"context"

	"github.com/google/uuid"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

var (
	// nil UUID for a platform-wide preference
	PlatformID = uuid.Nil.String()
)

type Repository interface {
	Set(ctx context.Context, preference Preference) (Preference, error)
	Get(ctx context.Context, id uuid.UUID) (Preference, error)
	List(ctx context.Context, filter Filter) ([]Preference, error)
}

type Service struct {
	repo   Repository
	traits []Trait
}

func NewService(repo Repository, traits []Trait) *Service {
	if traits == nil {
		traits = DefaultTraits
	}
	return &Service{
		repo:   repo,
		traits: traits,
	}
}

func (s *Service) Create(ctx context.Context, preference Preference) (Preference, error) {
	// only allow creating preferences for which a trait exists
	var matchedTrait *Trait
	for _, trait := range s.Describe(ctx) {
		if trait.Name == preference.Name && trait.ResourceType == preference.ResourceType {
			matchedTrait = &trait
			break
		}
	}
	if matchedTrait == nil {
		return Preference{}, ErrTraitNotFound
	}

	// validate scope
	hasScope := preference.ScopeType != "" || preference.ScopeID != ""
	traitRequiresScope := len(matchedTrait.AllowedScopes) > 0

	if traitRequiresScope {
		// Trait requires scope - must provide both scope_type and scope_id
		if preference.ScopeType == "" || preference.ScopeID == "" {
			return Preference{}, ErrInvalidScope
		}
		// Scope type must be in the trait's allowed scopes
		if !matchedTrait.IsValidScope(preference.ScopeType) {
			return Preference{}, ErrInvalidScope
		}
	} else if hasScope {
		// Trait is global-only but scope was provided
		return Preference{}, ErrInvalidScope
	}

	validator := matchedTrait.GetValidator()
	if !validator.Validate(preference.Value) {
		return Preference{}, ErrInvalidValue
	}
	created, err := s.repo.Set(ctx, preference)
	if err != nil {
		return Preference{}, err
	}
	// Populate ValueDescription from trait's InputOptions
	created.ValueDescription = matchedTrait.GetValueDescription(created.Value)
	return created, nil
}

func (s *Service) Get(ctx context.Context, id string) (Preference, error) {
	prefID, err := uuid.Parse(id)
	if err != nil {
		return Preference{}, ErrInvalidID
	}
	return s.repo.Get(ctx, prefID)
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Preference, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Describe(ctx context.Context) []Trait {
	return s.traits
}

// LoadUserPreferences loads user preferences and merges them with trait defaults.
// Always returns a complete preference set with priority:
//  1. Org-scoped DB values (if scope provided, highest priority)
//  2. Global DB values (fallback)
//  3. Trait defaults (for anything not in DB)
func (s *Service) LoadUserPreferences(ctx context.Context, filter Filter) ([]Preference, error) {
	hasScope := filter.ScopeType != "" && filter.ScopeID != ""

	// Fetch global preferences
	globalPrefs, err := s.repo.List(ctx, Filter{
		UserID: filter.UserID,
		// No scope = global preferences (repo will use ScopeTypeGlobal/ScopeIDGlobal)
	})
	if err != nil {
		return nil, err
	}

	// Build preference map starting with global preferences
	prefMap := make(map[string]Preference)
	for _, pref := range globalPrefs {
		prefMap[pref.Name] = pref
	}

	// If scope provided, fetch scoped preferences and override globals
	if hasScope {
		scopedPrefs, err := s.repo.List(ctx, filter)
		if err != nil {
			return nil, err
		}
		for _, pref := range scopedPrefs {
			prefMap[pref.Name] = pref
		}
	}

	// Build result with trait ordering, filling in defaults for missing preferences
	var result []Preference
	for _, trait := range s.traits {
		if trait.ResourceType != schema.UserPrincipal {
			continue
		}
		if pref, exists := prefMap[trait.Name]; exists {
			// Populate ValueDescription from trait's InputOptions
			pref.ValueDescription = trait.GetValueDescription(pref.Value)
			result = append(result, pref)
			delete(prefMap, trait.Name) // mark as processed
		} else if trait.Default != "" {
			// Add default preference for unset trait
			result = append(result, Preference{
				Name:             trait.Name,
				Value:            trait.Default,
				ValueDescription: trait.GetValueDescription(trait.Default),
				ResourceID:       filter.UserID,
				ResourceType:     schema.UserPrincipal,
				ScopeType:        filter.ScopeType,
				ScopeID:          filter.ScopeID,
			})
		}
	}

	// Add any remaining preferences that don't have a matching trait
	// (shouldn't happen normally but handles edge cases)
	for _, pref := range prefMap {
		result = append(result, pref)
	}

	return result, nil
}

// LoadPlatformPreferences loads platform preferences from the database
// and returns a map of preference name to value
// if a preference is not set in the database, the default value is used from DefaultTraits
func (s *Service) LoadPlatformPreferences(ctx context.Context) (map[string]string, error) {
	// TODO(kushsharma): we should cache this method as it will not happen that often

	preferences, err := s.List(ctx, Filter{
		ResourceID:   PlatformID,
		ResourceType: schema.PlatformNamespace,
	})
	if err != nil {
		return nil, err
	}

	// load platform config from preferences if set
	prefs := make(map[string]string)
	for _, pref := range preferences {
		prefs[pref.Name] = pref.Value
	}

	// load default platform config if not set in preferences already
	for _, t := range s.traits {
		if t.ResourceType == schema.PlatformNamespace && prefs[t.Name] == "" {
			prefs[t.Name] = t.Default
		}
	}
	return prefs, nil
}
