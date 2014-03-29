package makex

import "fmt"

// TargetsNeedingBuild returns an ordered list of target sets
func (c *Config) NewMaker(mf *Makefile) *Maker {
	m := &Maker{
		mf:     mf,
		Config: c,
	}
	m.buildDAG()
	return m
}

type Maker struct {
	mf     *Makefile
	dag    map[string][]string
	topo   [][]string
	cycles map[string][]string

	*Config
}

func (m *Maker) buildDAG() {
	// topological sort taken from
	// http://rosettacode.org/wiki/Topological_sort#Go.

	if m.dag == nil || m.cycles == nil {
		m.dag = make(map[string][]string)
		m.cycles = make(map[string][]string)
	}

	for _, rule := range m.mf.Rules {
		target := rule.Target()
		prereqs := m.dag[target] // handle additional dependencies

	scan:
		for _, pr := range rule.Prereqs() {
			for _, known := range prereqs {
				if known == pr {
					continue scan // ignore duplicate dependencies
				}
			}

			// make an node for the prereq target if it doesn't exist
			m.dag[pr] = m.dag[pr]

			// build: add edge (dependency)
			prereqs = append(prereqs, pr)
		}

		// add or update node for dependent target
		m.dag[target] = prereqs
	}

	// topological sort on the DAG
	for len(m.dag) > 0 {

		// collect targets with no dependencies
		var zero []string
		for target, prereqs := range m.dag {
			if len(prereqs) == 0 {
				zero = append(zero, target)
				delete(m.dag, target)
			}
		}

		// cycle detection
		if len(zero) == 0 {
			// collect un-orderable dependencies
			cycle := make(map[string]bool)
			for _, prereqs := range m.dag {
				for _, dep := range prereqs {
					cycle[dep] = true
				}
			}

			// mark targets with un-orderable dependencies
			for target, prereqs := range m.dag {
				if cycle[target] {
					m.cycles[target] = prereqs
				}
			}
			return
		}

		// output a set that can be processed concurrently
		m.topo = append(m.topo, zero)

		// remove edges (dependencies) from dg
		for _, remove := range zero {
			for target, prereqs := range m.dag {
				for i, dep := range prereqs {
					if dep == remove {
						copy(prereqs[i:], prereqs[i+1:])
						m.dag[target] = prereqs[:len(prereqs)-1]
						break
					}
				}
			}
		}
	}
}

func (m *Maker) TargetSets(goals []string) [][]string {
	return m.topo
}

// TODO!(sqs): use goals
func (m *Maker) TargetSetsNeedingBuild(goals ...string) ([][]string, error) {
	for _, goal := range goals {
		if rule := m.mf.Rule(goal); rule == nil {
			return nil, errNoRuleToMakeTarget(goal)
		}
		if deps, isCycle := m.cycles[goal]; isCycle {
			return nil, errCircularDependency(goal, deps)
		}
	}

	targetSets := make([][]string, 0)
	for _, targetSet := range m.topo {
		var targetsNeedingBuild []string
		for _, target := range targetSet {
			exists, err := m.pathExists(target)
			if err != nil {
				return nil, err
			}
			if !exists {
				targetsNeedingBuild = append(targetsNeedingBuild, target)
			}
		}
		if len(targetsNeedingBuild) > 0 {
			targetSets = append(targetSets, targetsNeedingBuild)
		}
	}
	return targetSets, nil
}

func errNoRuleToMakeTarget(target string) error {
	return fmt.Errorf("no rule to make target %q", target)
}

func errCircularDependency(target string, deps []string) error {
	return fmt.Errorf("circular dependency for target %q: %v", target, deps)
}
