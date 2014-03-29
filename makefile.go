package makex

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Makefile struct {
	Rules []Rule
}

type Rule interface {
	Target() string
	Prereqs() []string
	Recipes() []string
}

func Marshal(mf *Makefile) ([]byte, error) {
	var b bytes.Buffer

	var all []string
	for _, rule := range mf.Rules {
		ruleName := rule.Target()
		all = append(all, ruleName)
	}
	if len(all) > 0 {
		fmt.Fprintln(&b, ".PHONY: all")
		fmt.Fprintf(&b, "all: %s\n", strings.Join(all, " "))
	}

	for _, rule := range mf.Rules {
		fmt.Fprintln(&b)

		ruleName := rule.Target()
		fmt.Fprintf(&b, "%s:", ruleName)
		for _, prereq := range rule.Prereqs() {
			fmt.Fprintf(&b, " %s", prereq)
		}
		fmt.Fprintln(&b)
		for _, recipe := range rule.Recipes() {
			fmt.Fprintf(&b, "\t%s\n", recipe)
		}
	}

	return b.Bytes(), nil
}

func Make(dir string, makefile []byte, args []string) error {
	tmpFile, err := ioutil.TempFile("", "sg-makefile")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	err = ioutil.WriteFile(tmpFile.Name(), makefile, 0600)
	if err != nil {
		return err
	}

	args = append(args, "-f", tmpFile.Name(), "-C", dir)
	mk := exec.Command("make", args...)
	mk.Stdout = os.Stderr
	mk.Stderr = os.Stderr
	return mk.Run()
}

var cleanRE = regexp.MustCompile(`^[\w\d_/.-]+$`)

func Quote(s string) string {
	if cleanRE.MatchString(s) {
		return s
	}
	q := strconv.Quote(s)
	return "'" + strings.Replace(q[1:len(q)-1], "'", "", -1) + "'"
}
