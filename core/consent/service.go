package consent

import (
	"fmt"
	"sort"
	"strings"
)

// Service owns the document config, so it owns the checks that read it. Config
// is read at boot, so a version change needs a restart.
type Service struct {
	config Config
}

func NewService(config Config) *Service {
	return &Service{
		config: config,
	}
}

// Documents returns every configured document, ordered by id so a response built
// from it is stable. Empty when disabled, so one client build works against both
// kinds of deployment.
func (s Service) Documents() []Document {
	if !s.config.Enabled {
		return nil
	}

	ids := sortedIDs(s.config.Documents)
	documents := make([]Document, 0, len(ids))
	for _, id := range ids {
		documents = append(documents, s.document(id))
	}
	return documents
}

// Resolve maps ids to their config snapshots and rejects unknown ones, saying
// nothing about whether the set is complete — for callers where completeness is
// not yet knowable. Duplicates are removed.
//
// Disabled, it resolves nothing and rejects nothing, so the ids are ignored
// rather than refused.
func (s Service) Resolve(ids []string) ([]Document, error) {
	if !s.config.Enabled {
		return nil, nil
	}

	accepted := uniqueSorted(ids)
	var unknown []string
	for _, id := range accepted {
		if _, ok := s.config.Documents[id]; !ok {
			unknown = append(unknown, id)
		}
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrUnknownDocuments, strings.Join(unknown, ", "))
	}

	documents := make([]Document, 0, len(accepted))
	for _, id := range accepted {
		documents = append(documents, s.document(id))
	}
	return documents, nil
}

// ResolveAll is Resolve plus the completeness rule: the ids must cover every
// configured document, no more and no less. Both directions are compared, so the
// error names what is wrong rather than just that something is.
func (s Service) ResolveAll(ids []string) ([]Document, error) {
	if !s.config.Enabled {
		return nil, nil
	}

	// the "no more" direction and the deduplication both come from Resolve
	documents, err := s.Resolve(ids)
	if err != nil {
		return nil, err
	}

	accepted := make(map[string]struct{}, len(documents))
	for _, document := range documents {
		accepted[document.ID] = struct{}{}
	}

	var missing []string
	for _, id := range sortedIDs(s.config.Documents) {
		if _, ok := accepted[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrMissingDocuments, strings.Join(missing, ", "))
	}
	return documents, nil
}

func (s Service) document(id string) Document {
	document := s.config.Documents[id]
	return Document{
		ID:      id,
		Title:   document.Title,
		Version: document.Version,
		URL:     document.URL,
	}
}

// uniqueSorted so the same accepted set produces the same document list
// whatever order the client sent it in.
func uniqueSorted(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		seen[id] = struct{}{}
	}
	unique := make([]string, 0, len(seen))
	for id := range seen {
		unique = append(unique, id)
	}
	sort.Strings(unique)
	return unique
}
