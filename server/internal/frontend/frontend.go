package frontend

import "embed"

//go:embed index.html client.mjs
var EmbeddedFiles embed.FS
