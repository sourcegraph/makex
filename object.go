package makex

// An Object represents a target or a prerequisite.
type Object interface {
	Name() string
	Exists(c *Config) (bool, error)
}

type Filename string

func (o Filename) Name() string { return string(o) }

func (o Filename) Exists(c *Config) (bool, error) { return c.pathExists(string(o)) }

type Glob string

func (o Glob) Name() string { return string(o) }

func (o Glob) Exists(c *Config) (bool, error) { return c.pathExists(string(o)) }
