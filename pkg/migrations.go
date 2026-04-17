package foo

import "embed"

//go:embed migrations
var MigrationsFS embed.FS
