package makex

import (
	"bytes"
	"fmt"
	"strings"
)

// Parse parses a Makefile into a *Makefile struct.
//
// TODO(sqs): super hacky.
func Parse(data []byte) (*Makefile, error) {
	var mf Makefile

	lines := bytes.Split(data, []byte{'\n'})
	var rule *BasicRule
	for lineno, lineBytes := range lines {
		line := string(lineBytes)
		if strings.Contains(line, ":") {
			sep := strings.Index(line, ":")
			targets := strings.Fields(line[:sep])
			if len(targets) > 1 {
				return nil, errMultipleTargetsUnsupported(lineno)
			}
			target := targets[0]
			prereqs := strings.Fields(line[sep+1:])
			rule = &BasicRule{TargetFile: target, PrereqFiles: prereqs}
			mf.Rules = append(mf.Rules, rule)
		} else if strings.HasPrefix(line, "\t") {
			if rule == nil {
				return nil, fmt.Errorf("line %d: indented recipe not inside a rule", lineno)
			}
			recipe := strings.TrimPrefix(line, "\t")
			recipe = strings.Replace(recipe, "$@", Quote(rule.TargetFile), -1)
			recipe = strings.Replace(recipe, "$^", strings.Join(QuoteList(rule.PrereqFiles), " "), -1)
			rule.RecipeCmds = append(rule.RecipeCmds, recipe)
		} else {
			rule = nil
		}
	}

	return &mf, nil
}

func errMultipleTargetsUnsupported(lineno int) error {
	return fmt.Errorf("line %d: rule with multiple targets is yet implemented", lineno)
}
