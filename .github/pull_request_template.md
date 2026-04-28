## Summary
<!-- Provide a high-level overview of the changes in this PR -->


## Changes
<!-- List the key changes made in this PR -->
-


## Technical Details
<!-- Optional: Add implementation-specific details, architectural decisions, or technical context -->


## Test Plan
<!-- Describe how you tested these changes -->
- [ ] Manual testing completed
- [ ] Build and type checking passes

## SQL Safety (if your PR touches `*_repository.go` or `goqu.*`)
- [ ] Values flow through `?` placeholders, `goqu.Ex{}`, or `goqu.Record{}` — never `fmt.Sprintf` or `+` building a query that gets executed.
- [ ] `ToSQL()` callers capture and forward params (`query, params, err := stmt.ToSQL(); db.…Context(ctx, …, query, params...)`). Never `query, _, err := …`.
- [ ] No `?` placeholders inside single-quoted SQL literals in `goqu.L` (use `make_interval(hours => ?)`-style functions instead).
- [ ] Any `//nolint:forbidigo` or `// #nosec G20x` annotation has a one-line justification on the same line that a reviewer can verify.

