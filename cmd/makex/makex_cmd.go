package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/sourcegraph/makex"
)

var file = flag.String("f", "Makefile", "path to Makefile")
var cwd = flag.String("C", "", "change to this directory before doing anything")
var dryRun = flag.Bool("n", false, "dry run (don't actually run any commands)")

func main() {
	log := log.New(os.Stderr, "makex: ", 0)

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

	if *cwd != "" {
		err := os.Chdir(*cwd)
		if err != nil {
			log.Fatal(err)
		}
	}

	data, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}

	mf, err := makex.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	goals := flag.Args()
	if len(goals) == 0 && len(mf.Rules) > 0 {
		goals = []string{mf.Rules[0].Target()}
	}

	conf := makex.Default
	mk := conf.NewMaker(mf, goals...)
	targetSets, err := mk.TargetSetsNeedingBuild()
	if err != nil {
		log.Fatal(err)
	}

	if len(targetSets) == 0 {
		fmt.Println("Nothing to do.")
	}
	for _, targetSet := range targetSets {
		fmt.Println(targetSet)
	}
}
