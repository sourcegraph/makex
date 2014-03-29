package makex

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegraph/rwvfs"
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

// Expand returns a clone of mf with Prereqs filepath globs expanded. If rules
// contain globs, they are replaced with BasicRules with the globs expanded.
//
// Only globs containing "*" are detected.
func (c *Config) Expand(orig *Makefile) (*Makefile, error) {
	var mf Makefile
	mf.Rules = make([]Rule, len(orig.Rules))
	for i, rule := range orig.Rules {
		var hasGlob bool
		for _, target := range rule.Prereqs() {
			if strings.Contains(target, "*") {
				hasGlob = true
				break
			}
		}
		if hasGlob {
			var expandedPrereqs []string
			for _, target := range rule.Prereqs() {
				files, err := rwvfs.Glob(walkableRWVFS{c.fs()}, globPrefix(target), target)
				if err != nil {
					return nil, err
				}
				expandedPrereqs = append(expandedPrereqs, files...)
			}

			mf.Rules[i] = &BasicRule{
				TargetFile:  rule.Target(),
				PrereqFiles: expandedPrereqs,
				RecipeCmds:  rule.Recipes(),
			}
		} else {
			mf.Rules[i] = rule
		}
	}
	return &mf, nil
}

// globPrefix returns all path components up to (not including) the first path
// component that contains a "*".
func globPrefix(path string) string {
	cs := strings.Split(path, string(filepath.Separator))
	var prefix []string
	for _, c := range cs {
		if strings.Contains(c, "*") {
			break
		}
		prefix = append(prefix, c)
	}
	return filepath.Join(prefix...)
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
