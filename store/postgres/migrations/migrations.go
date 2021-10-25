package migrations

import "embed"

//go:embed *.sql
var MigrationFs embed.FS
