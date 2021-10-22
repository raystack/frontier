package embeds

import "embed"

//go:embed migrations/*
var MigrationFs embed.FS
