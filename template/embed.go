package template

import "embed"

//go:embed *.gitignore
var TemplatesFS embed.FS
