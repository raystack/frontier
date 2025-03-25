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
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
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
	validator := matchedTrait.GetValidator()
	if !validator.Validate(preference.Value) {
		return Preference{}, ErrInvalidValue
	}
	return s.repo.Set(ctx, preference)
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
	return DefaultTraits
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
	for _, t := range DefaultTraits {
		if t.ResourceType == schema.PlatformNamespace && prefs[t.Name] == "" {
			prefs[t.Name] = t.Default
		}
	}
	return prefs, nil
}
