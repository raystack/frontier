package schema_test

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const inheritanceTestTenant = "frontier"

func compileBaseSchema(t *testing.T) *compiler.CompiledSchema {
	t.Helper()
	compiled, err := compiler.Compile(compiler.InputSchema{
		Source:       "base_schema.zed",
		SchemaString: schema.BaseSchemaZed,
	}, compiler.ObjectTypePrefix(inheritanceTestTenant))
	require.NoError(t, err)
	return compiled
}

// TestExtractInheritance_Sanity asserts the canonical permissions which the
// schema embeds today actually show up in the extracted maps. If the schema
// changes, this test changes alongside it.
func TestExtractInheritance_Sanity(t *testing.T) {
	compiled := compileBaseSchema(t)
	got, err := schema.ExtractInheritance(compiled)
	require.NoError(t, err)

	t.Run("project direct visibility includes administer/get/update", func(t *testing.T) {
		assert.ElementsMatch(t, []string{
			"app_project_administer",
			"app_project_get",
			"app_project_update",
		}, got.ProjectDirectVisibility)
	})

	t.Run("org->project inheritance includes both granted and pat_granted arrows", func(t *testing.T) {
		assert.ElementsMatch(t, []string{
			"app_organization_administer",
			"app_project_get",
			"app_project_administer",
		}, got.OrganizationToProjectInherit)
	})
}

// TestExtractInheritance_OracleParity is the drift guard. A regex oracle reads
// the raw base_schema.zed source and computes the same lists by string parsing
// — if the AST walker ever misses an arrow (or picks up an extra one), the two
// disagree and this test fails.
//
// The oracle strips `//`-to-EOL comments and asserts each permission lives on
// a single line. If the schema ever wraps a permission across multiple lines,
// the assertion panics with a clear message so a human chooses how to update
// the oracle rather than letting the test silently produce wrong results.
func TestExtractInheritance_OracleParity(t *testing.T) {
	compiled := compileBaseSchema(t)
	got, err := schema.ExtractInheritance(compiled)
	require.NoError(t, err)

	projectGet := oracleArrows(t, schema.BaseSchemaZed, "app/project", "get", false)
	orgProjectGet := oracleArrows(t, schema.BaseSchemaZed, "app/organization", "project_get", true)

	assert.ElementsMatch(t, projectGet, got.ProjectDirectVisibility,
		"AST walker drifted from regex oracle for project.get")
	assert.ElementsMatch(t, orgProjectGet, got.OrganizationToProjectInherit,
		"AST walker drifted from regex oracle for org.project_get")
}

// oracleArrows is a deliberately dumb regex-based extractor: it locates the
// definition for `objectName`, finds the line beginning with `permission
// permissionName =`, and pulls out the relation names appearing after
// `granted->` (and `pat_granted->` if includePATGranted is set).
func oracleArrows(t *testing.T, source, objectName, permissionName string, includePATGranted bool) []string {
	t.Helper()
	body := definitionBody(t, source, objectName)

	commentRe := regexp.MustCompile(`//[^\n]*`)
	cleanedBody := commentRe.ReplaceAllString(body, "")

	permRe := regexp.MustCompile(`(?m)^\s*permission\s+` + regexp.QuoteMeta(permissionName) + `\s*=(.+)$`)
	matches := permRe.FindAllStringSubmatch(cleanedBody, -1)
	if len(matches) == 0 {
		t.Fatalf("oracle could not find `permission %s = ...` inside definition %q", permissionName, objectName)
	}
	if len(matches) > 1 {
		t.Fatalf("oracle expected exactly one `permission %s = ...` line inside definition %q, found %d", permissionName, objectName, len(matches))
	}
	// The regex anchors at line start/end with (?m). If the permission ever wraps
	// across lines, the captured group will be incomplete; panic to flag that
	// the oracle no longer matches the schema layout.
	expr := strings.TrimSpace(matches[0][1])
	if strings.HasSuffix(expr, "+") || strings.HasSuffix(expr, "&") || strings.HasSuffix(expr, "-") {
		t.Fatalf("oracle assumption violated: `permission %s` in %q appears to wrap across lines; update the oracle", permissionName, objectName)
	}

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
	sort.Strings(out)
	return out
}

// definitionBody returns the body of a `definition <objectName> { ... }` block
// from raw zed source. Matches by braces rather than indentation.
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

// TestExtractInheritance_RejectsNonUnion exercises the "fail loud" path: a
// permission defined with an intersection rather than a union must error so
// callers don't silently get the wrong inheritance list.
func TestExtractInheritance_RejectsNonUnion(t *testing.T) {
	src := `
definition app/role {
	relation app_project_get: app/user
}

definition app/rolebinding {
	relation bearer: app/user
	relation role: app/role

	permission app_project_get = bearer & role->app_project_get
}

definition app/project {
	relation org: app/organization
	relation granted: app/rolebinding

	permission get = granted->app_project_get & granted->app_project_get
}

definition app/organization {
	relation granted: app/rolebinding

	permission project_get = granted->app_project_get
}
`
	compiled, err := compiler.Compile(compiler.InputSchema{
		Source:       "test",
		SchemaString: src,
	}, compiler.ObjectTypePrefix(inheritanceTestTenant))
	require.NoError(t, err)

	_, err = schema.ExtractInheritance(compiled)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "intersection")
}
