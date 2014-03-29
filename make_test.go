package makex

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/rwvfs"
)

func TestTargetsNeedingBuild(t *testing.T) {
	tests := map[string]struct {
		mf                         *Makefile
		fs                         FileSystem
		goals                      []string
		wantErr                    error
		wantTargetSetsNeedingBuild [][]string
	}{
		"do nothing if empty": {
			mf: &Makefile{},
			wantTargetSetsNeedingBuild: [][]string{},
		},
		"return error if target isn't defined in Makefile": {
			mf:      &Makefile{},
			goals:   []string{"x"},
			wantErr: errNoRuleToMakeTarget("x"),
		},
		"don't build target that already exists": {
			mf:    &Makefile{Rules: []Rule{&BasicRule{TargetFile: "x"}}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{"x": ""})),
			goals: []string{"x"},
			wantTargetSetsNeedingBuild: [][]string{},
		},
		"build target that doesn't exist": {
			mf:    &Makefile{Rules: []Rule{&BasicRule{TargetFile: "x"}}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x"},
			wantTargetSetsNeedingBuild: [][]string{{"x"}},
		},
		"build targets recursively that don't exist": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"x1"}},
				&BasicRule{TargetFile: "x1"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0"},
			wantTargetSetsNeedingBuild: [][]string{{"x1"}, {"x0"}},
		},
		"don't build goal targets more than once": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0", "x0"},
			wantTargetSetsNeedingBuild: [][]string{{"x0"}},
		},
		"don't build any targets more than once": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "x1", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "y"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0", "x1"},
			wantTargetSetsNeedingBuild: [][]string{{"y"}, {"x0", "x1"}},
		},
		"detect 1-cycles": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"x0"}},
			}},
			fs:      NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:   []string{"x0"},
			wantErr: errCircularDependency("x0", []string{"x0"}),
		},
		"detect 2-cycles": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"x1"}},
				&BasicRule{TargetFile: "x1", PrereqFiles: []string{"x0"}},
			}},
			fs:      NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:   []string{"x0"},
			wantErr: errCircularDependency("x0", []string{"x1"}),
		},
	}

	for label, test := range tests {
		conf := &Config{FS: test.fs}
		mk := conf.NewMaker(test.mf)
		targetSets, err := mk.TargetSetsNeedingBuild(test.goals...)
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: TargetsNeedingBuild(%q): error: %s", label, test.goals, err)
				continue
			} else {
				t.Errorf("%s: TargetsNeedingBuild(%q): error: got %q, want %q", label, test.goals, err, test.wantErr)
				continue
			}
		}
		if !reflect.DeepEqual(targetSets, test.wantTargetSetsNeedingBuild) {
			t.Errorf("%s: got targetSets needing build %v, want %v", label, targetSets, test.wantTargetSetsNeedingBuild)
		}
	}
}
