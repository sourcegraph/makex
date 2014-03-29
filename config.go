package makex

import (
	"os"

	"github.com/sourcegraph/rwvfs"
)

type Config struct {
	FS FileSystem
}

var Default Config

func (c *Config) fs() FileSystem {
	if c.FS != nil {
		return c.FS
	}
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	return NewFileSystem(rwvfs.OS(dir))
}
