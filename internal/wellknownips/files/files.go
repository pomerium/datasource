package files

import (
	"embed"
)

//go:embed azure.json.gz
var FS embed.FS
