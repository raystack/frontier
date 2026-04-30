# Project Conventions for Agents and Contributors

This directory holds rules an agent (or human) should follow when working on Frontier. Each file is short and prescriptive — read the ones relevant to your task.

## Layout

- [`backend/`](backend/) — Go services in `cmd/`, `core/`, `internal/`, `billing/`, `pkg/`
- [`frontend/`](frontend/) — Web monorepo under `web/` (pnpm + Turbo, React + TS)

## Index

### Backend
- [SQL Safety](backend/sql-safety.md) — DB query rules: prepared statements, params handling, `goqu` pitfalls.

### Frontend
- _(none yet)_

## Adding new conventions

Drop a markdown file in the relevant subfolder and link it from this README. Keep each file short and rule-focused — "what to do, why, how lint enforces it." Long explanations belong in design docs or PR descriptions, not here.