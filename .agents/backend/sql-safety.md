# SQL Safety (Backend)

All DB queries use [goqu](https://github.com/doug-martin/goqu) for SQL building over [sqlx](https://github.com/jmoiron/sqlx)/`database/sql`, driver `pgx`. Prepared statements are forced globally via `goqu.SetDefaultPrepared(true)` in `cmd/serve.go::setupDB` — don't disable.

Lint enforces part of these rules (`.golangci.yml`: `gosec` G201/G202 + two `forbidigo` patterns). Review enforces the rest.

## The core rule

User-supplied input that reaches the DB is **always parameterized**. The SQL template is fixed; values are bound separately. Assume any value from an HTTP request, gRPC message, or external system is hostile.

```go
// good: value bound as $1
dialect.From("users").Where(goqu.Ex{"email": email}).ToSQL()

// bad: value spliced into the SQL string
fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
```

## Specific rules

### 1. Use `goqu.Ex{}` / `goqu.Record{}` for values, never `fmt.Sprintf` or `+`

```go
// good
dialect.Insert("users").Rows(goqu.Record{"email": email, "name": name})
dialect.From("users").Where(goqu.Ex{"id": id})

// bad
db.Exec(fmt.Sprintf("INSERT INTO users (email) VALUES ('%s')", email))
db.Exec("DELETE FROM users WHERE id = '" + id + "'")
```

Both bad forms are flagged by `gosec G201/G202`.

### 2. Capture and forward `params` from `ToSQL()`

```go
// good
query, params, err := stmt.ToSQL()
db.SelectContext(ctx, &out, query, params...)

// bad — silently broken with prepared mode: SQL has $N placeholders but no values flow
query, _, err := stmt.ToSQL()
db.SelectContext(ctx, &out, query)
```

The `query, _, err := …ToSQL()` shape is flagged by a `forbidigo` rule.

### 3. Don't put `?` inside single-quoted SQL strings in `goqu.L`

```go
// bad — goqu replaces ? regardless of quote context; breaks in prepared mode
goqu.L("expires_at + INTERVAL '? hours'", n)

// good — use a Postgres function whose args are bindable
goqu.L("expires_at + make_interval(secs => ?)", duration.Seconds())
```

Flagged by a `forbidigo` rule.

> Note: `make_interval` takes `years/months/weeks/days/hours/mins` as `int` and only `secs` as `double precision`. If sub-unit precision matters (e.g., a `time.Duration` that may have minutes/seconds), use `secs => ?` with `duration.Seconds()` rather than `hours => ?` with `int(duration.Hours())` to avoid silent truncation.

### 4. Always start a query with `dialect.From(…)`, never `goqu.From(…)`

```go
// good — uses the postgres dialect ($N placeholders)
dialect.From(TABLE_USERS).Where(...)

// bad — uses goqu's default dialect (? placeholders) which PostgreSQL rejects
goqu.From("users").Where(...)
```

With prepared mode on, `goqu.From` produces `?` and Postgres errors with `syntax error at or near ")"`. Flagged by a `forbidigo` rule.

### 5. Identifier splicing (table/column names) uses `goqu.I` / `goqu.T`

```go
// good
goqu.I("users.email")
goqu.T("users").As("u")

// avoid (works for compile-time constants, but error-prone if a variable sneaks in)
goqu.L("users." + columnName)
```

Not lint-enforced. Reviewer catches any new `goqu.L` first-arg that could carry user input.

### 6. When you must use `fmt.Sprintf`, justify it on the same line

Postgres rejects `$N` parameter binding for cursor names, table/column names, `FETCH`/`LIMIT` counts in some positions — `fmt.Sprintf` is the only option. When you do this:

- Every spliced value must be server-controlled (compile-time constant, server-generated identifier, etc.) — not user input.
- Suppress the lint with a one-line justification on the same line:

```go
// #nosec G201 -- cursorName is server-generated (crypto/rand hex); Postgres
// has no parameter binding for cursor names.
fetchSQL := fmt.Sprintf("FETCH %d FROM %s", BATCH_SIZE, cursorName)
```

The justification must let a reviewer verify the claim without leaving the file.

## Lint coverage at a glance

| Rule | Source | Catches |
|---|---|---|
| G201 | `gosec` | `fmt.Sprintf` whose result is passed to `db.Query`/`db.Exec` |
| G202 | `gosec` | string concatenation passed to `db.Query`/`db.Exec` |
| `?` in `'…'` | `forbidigo` | `goqu.L` with `?` inside single quotes (single-line regex; line-wrapped literals are not detected) |
| discard params | `forbidigo` | `query, _, err := stmt.ToSQL()` (single-line regex; multi-line `query, _,\n\terr := …ToSQL()` and `query, _, _ := …ToSQL()` are not detected) |
| `goqu.From(` | `forbidigo` | use `dialect.From(...)` instead — the package-level `goqu.From` defaults to a non-postgres dialect with `?` placeholders |

Code-smell patterns (`goqu.L(fmt.Sprintf(…))`, `goqu.L("…" + var + …)`) are intentionally **not** lint-enforced — they fire too often on safe compile-time `TABLE_*`/`COLUMN_*` constant splicing. PR review catches the few cases where a variable in those positions could carry user input.

## When lint fires

Prefer fixing the code shape over annotating. Annotate (`// #nosec G20x` or `//nolint:forbidigo`) only when there's no parameterizable alternative — e.g., Postgres can't bind that position. Always include a one-line reason a reviewer can verify.