package makex

import (
	"fmt"
	"os"
)

// TargetsNeedingBuild returns a slice of target names that are outdated or
// nonexistent.
func (c *Config) TargetsNeedingBuild(mf *Makefile, goals ...string) ([]string, error) {
	fs := c.fs()

	targets := make([]string, 0)
	for _, goal := range goals {
		rule := mf.Rule(goal)
		if rule == nil {
			return nil, errNoRuleToMakeTarget(goal)
		}

		_, err := fs.Stat(goal)
		if os.IsNotExist(err) {
			prereqTargets, err := c.TargetsNeedingBuild(mf, rule.Prereqs()...)
			if err != nil {
				return nil, err
			}
			targets = append(targets, prereqTargets...)
			targets = append(targets, goal)
		} else if err != nil {
			return nil, err
		}
	}
	return targets, nil
}

func errNoRuleToMakeTarget(target string) error {
	return fmt.Errorf("no rule to make target %q", target)
}
