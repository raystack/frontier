package adminui

import (
	"embed"
)

//go:embed all:dist/admin-ui
var Assets embed.FS
