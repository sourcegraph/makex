package makex

import (
	"log"
	"os"

	"github.com/sourcegraph/rwvfs"
)

type Config struct {
	FS           FileSystem
	ParallelJobs int
	Log          *log.Logger
}

var Default = Config{
	ParallelJobs: 1,
	Log:          log.New(os.Stderr, "makex: ", 0),
}

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

func (c *Config) pathExists(path string) (bool, error) {
	_, err := c.fs().Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
