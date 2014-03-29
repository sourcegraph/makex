package makex

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/rwvfs"
)

func TestTargetsNeedingBuild(t *testing.T) {
	tests := map[string]struct {
		mf             *Makefile
		fs             FileSystem
		goals          []string
		wantErr        error
		wantRulesToRun []string
	}{
		"do nothing if empty": {
			mf:             &Makefile{},
			wantRulesToRun: []string{},
		},
		"return error if target isn't defined in Makefile": {
			mf:      &Makefile{},
			goals:   []string{"x"},
			wantErr: errNoRuleToMakeTarget("x"),
		},
		"don't build target that already exists": {
			mf:             &Makefile{Rules: []Rule{dummyRule{target: "x"}}},
			fs:             NewFileSystem(rwvfs.Map(map[string]string{"x": ""})),
			goals:          []string{"x"},
			wantRulesToRun: []string{},
		},
		"build target that doesn't exist": {
			mf:             &Makefile{Rules: []Rule{dummyRule{target: "x"}}},
			fs:             NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:          []string{"x"},
			wantRulesToRun: []string{"x"},
		},
		"build targets recursively that don't exist": {
			mf: &Makefile{Rules: []Rule{
				dummyRule{target: "x0", prereqs: []string{"x1"}},
				dummyRule{target: "x1"},
			}},
			fs:             NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:          []string{"x0"},
			wantRulesToRun: []string{"x1", "x0"},
		},
		"don't build goal targets more than once": {
			mf: &Makefile{Rules: []Rule{
				dummyRule{target: "x0"},
			}},
			fs:             NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:          []string{"x0", "x0"},
			wantRulesToRun: []string{"x0"},
		},
		"don't build any targets more than once": {
			mf: &Makefile{Rules: []Rule{
				dummyRule{target: "x0", prereqs: []string{"y"}},
				dummyRule{target: "x1", prereqs: []string{"y"}},
				dummyRule{target: "y"},
			}},
			fs:             NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:          []string{"x0", "x1"},
			wantRulesToRun: []string{"y", "x0", "x1"},
		},
		"detect 1-cycles": {
			mf: &Makefile{Rules: []Rule{
				dummyRule{target: "x0", prereqs: []string{"x0"}},
			}},
			fs:      NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:   []string{"x0"},
			wantErr: errCircularDependency("x0"),
		},
		"detect 2-cycles": {
			mf: &Makefile{Rules: []Rule{
				dummyRule{target: "x0", prereqs: []string{"x1"}},
				dummyRule{target: "x1", prereqs: []string{"x0"}},
			}},
			fs:      NewFileSystem(rwvfs.Map(map[string]string{})),
			goals:   []string{"x0"},
			wantErr: errCircularDependency("x0", "x1"),
		},
	}

	for label, test := range tests {
		conf := &Config{FS: test.fs}
		rulesToRun, err := conf.TargetsNeedingBuild(test.mf, test.goals...)
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: TargetsNeedingBuild(%q): error: %s", label, test.goals, err)
				continue
			} else {
				t.Errorf("%s: TargetsNeedingBuild(%q): error: got %q, want %q", label, test.goals, err, test.wantErr)
				continue
			}
		}
		if !reflect.DeepEqual(rulesToRun, test.wantRulesToRun) {
			t.Errorf("%s: got rulesToRun %v, want %v", label, rulesToRun, test.wantRulesToRun)
		}
	}
}
