package admin

import (
	"embed"
)

//go:embed all:dist
var Assets embed.FS
