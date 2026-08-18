package web

import "embed"

// Dist holds the production frontend. Run `make frontend` to populate dist/.
//
//go:embed all:dist
var Dist embed.FS
