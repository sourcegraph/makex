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

	fd := makex.Flags(nil, "")
	flag.Parse()
	conf := makex.Default
	fd.SetConfig(&conf)

	data, err := ioutil.ReadFile(fd.Makefile)
	if err != nil {
		conf.Log.Fatal(err)
	}

	if fd.Dir != "" {
		err := os.Chdir(fd.Dir)
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

	if fd.Expand {
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

	if fd.DryRun {
		for i, targetSet := range targetSets {
			if i != 0 {
				fmt.Println()
			}
			fmt.Printf("========= TARGET SET %d (%d targets)\n", i, len(targetSet))
			for _, target := range targetSet {
				fmt.Println(" - ", target)
			}
		}
		return
	}

	err = mk.Run()
	if err != nil {
		conf.Log.Fatal(err)
	}
}
