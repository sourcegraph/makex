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
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `makex is an experimental, incomplete implementation of make in Go.

Usage:

        makex [options] [target] ...

The options are:
`)
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()
	goals := flag.Args()

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

	conf := makex.Default
	mk := conf.NewMaker(mf)
	targetSets, err := mk.TargetSetsNeedingBuild(goals...)
	if err != nil {
		log.Fatal(err)
	}

	for _, targetSet := range targetSets {
		fmt.Println(targetSet)
	}
}
