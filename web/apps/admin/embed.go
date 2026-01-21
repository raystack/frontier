package adminui

import (
	"embed"
)

//go:embed all:dist/admin
var Assets embed.FS
