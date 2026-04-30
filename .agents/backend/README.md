# Backend Conventions

Rules for the Go backend (`cmd/`, `core/`, `internal/`, `billing/`, `pkg/`).

## Files

- [`sql-safety.md`](sql-safety.md) — DB query rules: prepared statements, `ToSQL()` params handling, `goqu.L` pitfalls, when `fmt.Sprintf` is unavoidable.

## Adding a backend convention

Create a new `.md` file in this directory, link it above. Keep each file focused on one topic. If a rule is enforced by lint, mention which lint rule (file + linter name) so a reader can find it.