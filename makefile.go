package makex

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Makefile struct {
	Rules []Rule
}

type BasicRule struct {
	TargetFile  string
	PrereqFiles []string
	RecipeCmds  []string
}

func (r *BasicRule) Target() string    { return r.TargetFile }
func (r *BasicRule) Prereqs() []string { return r.PrereqFiles }
func (r *BasicRule) Recipes() []string { return r.RecipeCmds }

// Rule returns the rule to make the specified target if it exists, or nil
// otherwise.
//
// TODO(sqs): support multiple rules for one target
// (http://www.gnu.org/software/make/manual/html_node/Multiple-Rules.html).
func (mf *Makefile) Rule(target string) Rule {
	for _, rule := range mf.Rules {
		if rule.Target() == target {
			return rule
		}
	}
	return nil
}

type Rule interface {
	Target() string
	Prereqs() []string
	Recipes() []string
}

func Marshal(mf *Makefile) ([]byte, error) {
	return marshal(mf, true)
}

func marshal(mf *Makefile, allRule bool) ([]byte, error) {
	var b bytes.Buffer

	if allRule {
		var all []string
		for _, rule := range mf.Rules {
			ruleName := rule.Target()
			all = append(all, ruleName)
		}
		if len(all) > 0 {
			fmt.Fprintln(&b, ".PHONY: all")
			fmt.Fprintf(&b, "all: %s\n", strings.Join(all, " "))
		}
		fmt.Fprintln(&b)
	}

	for i, rule := range mf.Rules {
		if i != 0 {
			fmt.Fprintln(&b)
		}

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

var cleanRE = regexp.MustCompile(`^[\w\d_/.-]+$`)

func Quote(s string) string {
	if cleanRE.MatchString(s) {
		return s
	}
	q := strconv.Quote(s)
	return "'" + strings.Replace(q[1:len(q)-1], "'", "", -1) + "'"
}

func QuoteList(ss []string) []string {
	q := make([]string, len(ss))
	for i, s := range ss {
		q[i] = Quote(s)
	}
	return q
}
