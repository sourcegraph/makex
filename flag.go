package makex

import (
	"flag"
	"runtime"
)

type FlagData struct {
	Makefile     string
	Dir          string
	DryRun       bool
	ParallelJobs int
	Expand       bool
	Verbose      bool
}

func (d *FlagData) SetConfig(c *Config) {
	c.ParallelJobs = d.ParallelJobs
	c.Verbose = d.Verbose
}

// Flags adds makex command-line flags to an existing flag.FlagSet (or the
// global FlagSet if fs is nil).
func Flags(fs *flag.FlagSet, prefix string) *FlagData {
	if fs == nil {
		fs = flag.CommandLine
	}
	var d FlagData
	fs.StringVar(&d.Makefile, prefix+"f", "Makefile", "path to Makefile")
	fs.StringVar(&d.Dir, prefix+"C", "", "change to this directory before doing anything")
	fs.BoolVar(&d.DryRun, prefix+"n", false, "dry run (don't actually run any commands)")
	fs.IntVar(&d.ParallelJobs, prefix+"j", runtime.GOMAXPROCS(0), "number of jobs to run in parallel")
	fs.BoolVar(&d.Expand, prefix+"x", true, "expand globs in makefile prereqs")
	fs.BoolVar(&d.Verbose, prefix+"v", false, "verbose")
	return &d
}
