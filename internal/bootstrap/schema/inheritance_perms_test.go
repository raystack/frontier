// Drift guard for the inheritance perm constants: regex-scans base_schema.zed
// and asserts ProjectDirectVisibilityPerms and OrganizationProjectInheritPerms
// match the granted-> / pat_granted-> arrows they're supposed to mirror.
package schema_test

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInheritancePerms_MatchSchema(t *testing.T) {
	cases := []struct {
		name              string
		objectName        string
		permissionName    string
		includePATGranted bool
		want              []string
	}{
		{
			name:              "ProjectDirectVisibilityPerms matches app/project.get granted-> arrows",
			objectName:        "app/project",
			permissionName:    "get",
			includePATGranted: false,
			want:              schema.ProjectDirectVisibilityPerms,
		},
		{
			name:              "OrganizationProjectInheritPerms matches app/organization.project_get granted-> and pat_granted-> arrows",
			objectName:        "app/organization",
			permissionName:    "project_get",
			includePATGranted: true,
			want:              schema.OrganizationProjectInheritPerms,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := arrowsFromSchemaSource(t, schema.BaseSchemaZed, tc.objectName, tc.permissionName, tc.includePATGranted)
			want := append([]string(nil), tc.want...)
			sort.Strings(got)
			sort.Strings(want)
			assert.Equal(t, want, got,
				"drifted from %s.%s in base_schema.zed; update the constant or the schema",
				tc.objectName, tc.permissionName)
		})
	}
}

// arrowsFromSchemaSource finds `permission <name> = ...` inside `definition
// <object> { ... }` and returns the deduped relation names on its granted->
// (and optionally pat_granted->) arrows.
func arrowsFromSchemaSource(t *testing.T, source, objectName, permissionName string, includePATGranted bool) []string {
	t.Helper()
	body := definitionBody(t, source, objectName)
	body = regexp.MustCompile(`//[^\n]*`).ReplaceAllString(body, "")

	permRe := regexp.MustCompile(`(?m)^\s*permission\s+` + regexp.QuoteMeta(permissionName) + `\s*=(.+)$`)
	matches := permRe.FindAllStringSubmatch(body, -1)
	require.Len(t, matches, 1, "expected exactly one `permission %s = ...` line in %q", permissionName, objectName)

	expr := strings.TrimSpace(matches[0][1])
	require.False(t,
		strings.HasSuffix(expr, "+") || strings.HasSuffix(expr, "&") || strings.HasSuffix(expr, "-"),
		"oracle assumption broken: `permission %s` in %q wraps across lines — rewrite the regex",
		permissionName, objectName)
	// Only Union (+) matches filterByRolePermissions's any-of semantics. Strip
	// arrows first so the `-` in `->` doesn't trip the check.
	opsOnly := strings.ReplaceAll(expr, "->", "")
	require.False(t,
		strings.ContainsAny(opsOnly, "&-"),
		"`permission %s` in %q uses intersection/exclusion — extend the gating logic before updating the constants",
		permissionName, objectName)

	arrowRe := regexp.MustCompile(`(granted|pat_granted)->([A-Za-z0-9_]+)`)
	out := make([]string, 0)
	seen := make(map[string]struct{})
	for _, m := range arrowRe.FindAllStringSubmatch(expr, -1) {
		kind, name := m[1], m[2]
		if kind == "pat_granted" && !includePATGranted {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// definitionBody returns the body of `definition <objectName> { ... }` by
// brace-walking — robust to nested braces if the schema ever grows them.
func definitionBody(t *testing.T, source, objectName string) string {
	t.Helper()
	header := "definition " + objectName + " {"
	start := strings.Index(source, header)
	require.GreaterOrEqual(t, start, 0, "definition %q not found in source", objectName)
	open := start + len(header) - 1
	depth := 0
	for i := open; i < len(source); i++ {
		switch source[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return source[open+1 : i]
			}
		}
	}
	t.Fatalf("definition %q has no matching closing brace", objectName)
	return ""
}
