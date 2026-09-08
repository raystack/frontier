package consent

import (
	"fmt"
	"net/url"
	"sort"
)

// Config lists the documents a deployment asks people to accept before an
// account is created. It sits at app.consent, beside app.authentication.
//
// Keyed by document id rather than a list, matching how authenticate.Config
// keys oidc_config: the key enforces unique ids and stays env-overridable.
// Every document is required at signup, so there is no per-document flag.
type Config struct {
	// Enabled switches the whole feature off by default.
	Enabled bool `yaml:"enabled" mapstructure:"enabled" default:"false"`

	Documents map[string]DocumentConfig `yaml:"documents" mapstructure:"documents"`
}

type DocumentConfig struct {
	Title string `yaml:"title" mapstructure:"title"`
	// Version is copied into the consent record and compared for equality only.
	Version string `yaml:"version" mapstructure:"version"`
	URL     string `yaml:"url" mapstructure:"url"`
}

// Validate runs at boot, so bad config stops the server rather than surfacing
// on someone's signup. A disabled block is not checked at all: nothing reads it,
// so a half-written map is only an error once consent is turned on.
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	// failing here rather than silently disabling itself, which would look
	// identical to a working deployment while asking nobody to accept anything
	if len(c.Documents) == 0 {
		return fmt.Errorf("app.consent is enabled but configures no documents")
	}

	// sorted so the error names the same document on every boot
	for _, id := range sortedIDs(c.Documents) {
		doc := c.Documents[id]
		if id == "" {
			return fmt.Errorf("app.consent has a document with an empty id")
		}
		// the title is what a client labels the link with, so a document without
		// one renders as a bare url the user is asked to accept
		if doc.Title == "" {
			return fmt.Errorf("app.consent document %q has an empty title", id)
		}
		if doc.Version == "" {
			return fmt.Errorf("app.consent document %q has an empty version", id)
		}
		if doc.URL == "" {
			return fmt.Errorf("app.consent document %q has an empty url", id)
		}
		parsed, err := url.Parse(doc.URL)
		if err != nil {
			return fmt.Errorf("app.consent document %q has an unparseable url %q: %w", id, doc.URL, err)
		}
		// url.Parse accepts a bare path, which no client can link to
		if !parsed.IsAbs() || parsed.Host == "" {
			return fmt.Errorf("app.consent document %q needs an absolute url with a host, got %q", id, doc.URL)
		}
	}
	return nil
}

func sortedIDs(documents map[string]DocumentConfig) []string {
	ids := make([]string, 0, len(documents))
	for id := range documents {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
