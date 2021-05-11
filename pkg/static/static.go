package static

import (
	"embed"
)

// StaticFS contains the static files
//go:embed files
var StaticFS embed.FS
