package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/sourcegraph/makex"
)

var file = flag.String("f", "Makefile", "path to Makefile")
var cwd = flag.String("C", "", "change to this directory before doing anything")
var dryRun = flag.Bool("n", false, "dry run (don't actually run any commands)")
var jobs = flag.Int("j", runtime.GOMAXPROCS(0), "number of jobs to run in parallel")

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `makex is an experimental, incomplete implementation of make in Go.

Usage:

        makex [options] [target] ...

If no targets are specified, the first target that appears in the makefile is
used.

The options are:
`)
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()
	conf := makex.Default
	conf.ParallelJobs = *jobs

	data, err := ioutil.ReadFile(*file)
	if err != nil {
		conf.Log.Fatal(err)
	}

	if *cwd != "" {
		err := os.Chdir(*cwd)
		if err != nil {
			conf.Log.Fatal(err)
		}
	}

	mf, err := makex.Parse(data)
	if err != nil {
		conf.Log.Fatal(err)
	}

	goals := flag.Args()
	if len(goals) == 0 {
		// Find the first rule that doesn't begin with a ".".
		for _, rule := range mf.Rules {
			target := rule.Target()
			if !strings.HasPrefix(target, ".") {
				goals = []string{target}
				break
			}
		}
	}

	mk := conf.NewMaker(mf, goals...)

	targetSets, err := mk.TargetSetsNeedingBuild()
	if err != nil {
		conf.Log.Fatal(err)
	}

	if len(targetSets) == 0 {
		fmt.Println("Nothing to do.")
	}

	if *dryRun {
		for _, targetSet := range targetSets {
			fmt.Println(targetSet)
		}
		return
	}

	err = mk.Run()
	if err != nil {
		conf.Log.Fatal(err)
	}
}
