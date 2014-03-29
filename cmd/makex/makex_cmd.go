package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/sourcegraph/makex"
)

var expand = flag.Bool("x", true, "expand globs in makefile prereqs")
var cwd = flag.String("C", "", "change to this directory before doing anything")
var file = flag.String("f", "Makefile", "path to Makefile")

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `makex is an experimental, incomplete implementation of make in Go.

Usage:

        makex [options] [target] ...

If no targets are specified, the first target that appears in the makefile (not
beginning with ".") is used.

The options are:
`)
		flag.PrintDefaults()
		os.Exit(1)
	}

	conf := makex.Default
	makex.Flags(nil, &conf, "")
	flag.Parse()

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

	if *expand {
		mf, err = conf.Expand(mf)
		if err != nil {
			log.Fatal(err)
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

	if conf.DryRun {
		mk.DryRun(os.Stdout)
		return
	}

	err = mk.Run()
	if err != nil {
		conf.Log.Fatal(err)
	}
}
