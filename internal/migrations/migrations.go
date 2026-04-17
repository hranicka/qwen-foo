package migrations

import (
	"embed"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// NewDriver returns a migrate source driver using the given embedded filesystem.
func NewDriver(fs embed.FS) (source.Driver, error) {
	return iofs.New(fs, "migrations")
}
