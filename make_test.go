package makex

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/rwvfs"
)

func TestMaker_DryRun(t *testing.T) {
	var conf Config
	mf := &Makefile{
		Rules: []Rule{&BasicRule{TargetFile: "x"}},
	}
	mk := conf.NewMaker(mf, "x")
	err := mk.DryRun(ioutil.Discard)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMaker_Run(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "makex")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	conf := &Config{
		ParallelJobs: 1,
		FS:           NewFileSystem(rwvfs.OS(tmpDir)),
	}

	target := "x"
	mf := &Makefile{
		Rules: []Rule{
			&BasicRule{
				TargetFile: target,
				RecipeCmds: []string{"touch " + filepath.ToSlash(filepath.Join(tmpDir, target))},
			},
		},
	}

	if isFile(conf.FS, target) {
		t.Fatalf("target %s exists before running Makefile; want it to not exist yet", target)
	}

	mk := conf.NewMaker(mf, target)
	err = mk.Run()
	if err != nil {
		t.Fatalf("Run failed: %s", err)
	}

	if !isFile(conf.FS, target) {
		t.Fatalf("target %s does not exist after running Makefile; want it to exist", target)
	}
}

func isFile(fs rwvfs.FileSystem, file string) bool {
	fi, err := fs.Stat(file)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

func newModTimeFileSystem(fs rwvfs.FileSystem) FileSystem {
	return modTimeFileSystem{walkableRWVFS{fs}, map[string]time.Time{}}
}

// modTimeFileSystem stores and retrieves mtimes independently of
// walkableRWVFS. For a new modTimeFileSystem, all of the mtimes are
// the zero value for time.Time. If a file is created and closed, the
// current time is stored as its mtime.
type modTimeFileSystem struct {
	walkableRWVFS
	modTimes map[string]time.Time
}

type modTimeFileInfo struct {
	modTime time.Time
	name    string
	mode    os.FileMode
	size    int64
	dir     bool
}

func (fi modTimeFileInfo) IsDir() bool        { return fi.dir }
func (fi modTimeFileInfo) ModTime() time.Time { return fi.modTime }
func (fi modTimeFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi modTimeFileInfo) Name() string       { return fi.name }
func (fi modTimeFileInfo) Size() int64        { return fi.size }
func (fi modTimeFileInfo) Sys() interface{}   { return nil }

func (m modTimeFileSystem) Stat(path string) (os.FileInfo, error) {
	fi, err := m.walkableRWVFS.Stat(path)
	if err != nil {
		return nil, err
	}
	return modTimeFileInfo{
		modTime: m.modTimes[filepath.Clean(path)],
		name:    fi.Name(),
		mode:    fi.Mode(),
		size:    fi.Size(),
		dir:     fi.IsDir(),
	}, nil
}

func (m modTimeFileSystem) Lstat(path string) (os.FileInfo, error) {
	fi, err := m.walkableRWVFS.Lstat(path)
	if err != nil {
		return nil, err
	}
	return modTimeFileInfo{
		modTime: m.modTimes[filepath.Clean(path)],
		name:    fi.Name(),
		mode:    fi.Mode(),
		size:    fi.Size(),
		dir:     fi.IsDir(),
	}, nil
}

func (m modTimeFileSystem) Create(path string) (io.WriteCloser, error) {
	f, err := m.walkableRWVFS.Create(path)
	return modTimeFile{f, path, m.modTimes}, err
}

type modTimeFile struct {
	io.WriteCloser
	path     string
	modTimes map[string]time.Time
}

func (m modTimeFile) Close() error {
	m.modTimes[filepath.Clean(m.path)] = time.Now()
	return m.WriteCloser.Close()
}

func TestTargetsNeedingBuild(t *testing.T) {
	tests := map[string]struct {
		mf    *Makefile
		fs    FileSystem
		goals []string
		// If afterMake is set, the test will run 'make' once,
		// call afterMake with fs, and then check the test
		// conditions.
		afterMake                  func(fs FileSystem) error
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
		"build only target with stale prereq": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x", PrereqFiles: []string{"x1"}},
				&BasicRule{TargetFile: "y", PrereqFiles: []string{"y1"}},
			}},
			fs: newModTimeFileSystem(rwvfs.Map(map[string]string{
				"x": "", "x1": "", "y": "", "y1": "",
			})),
			afterMake: func(fs FileSystem) error {
				w, err := fs.Create("x1")
				if err != nil {
					return err
				}
				defer w.Close()
				if _, err := io.WriteString(w, "modified"); err != nil {
					return err
				}
				return nil
			},
			goals: []string{"x", "y"},
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

		"don't build targets that don't directly achieve goals (simple)": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0"},
				&BasicRule{TargetFile: "x1"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0"},
			wantTargetSetsNeedingBuild: [][]string{{"x0"}},
		},
		"don't build targets that don't achieve goals (complex)": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "x1"},
				&BasicRule{TargetFile: "y"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0"},
			wantTargetSetsNeedingBuild: [][]string{{"y"}, {"x0"}},
		},
		"don't build targets that don't achieve goals (even when a common prereq is satisfied)": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: "x0", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "x1", PrereqFiles: []string{"y"}},
				&BasicRule{TargetFile: "y"},
			}},
			fs:    NewFileSystem(rwvfs.Map(map[string]string{})),
			goals: []string{"x0"},
			wantTargetSetsNeedingBuild: [][]string{{"y"}, {"x0"}},
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
		"re-build .PHONY target": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: ".PHONY", PrereqFiles: []string{"all"}},
				&BasicRule{TargetFile: "all", PrereqFiles: []string{"file"}},
			}},
			fs: newModTimeFileSystem(rwvfs.Map(map[string]string{
				"all": "", "file": "",
			})),
			goals: []string{"all"},
			wantTargetSetsNeedingBuild: [][]string{{"all"}},
		},
		"re-build .PHONY pre-requisite": {
			mf: &Makefile{Rules: []Rule{
				&BasicRule{TargetFile: ".PHONY", PrereqFiles: []string{"all", "compile"}},
				&BasicRule{TargetFile: "all", PrereqFiles: []string{"compile"}},
				&BasicRule{TargetFile: "compile", PrereqFiles: []string{"file"}},
			}},
			fs: newModTimeFileSystem(rwvfs.Map(map[string]string{
				"all": "", "compile": "", "file": "",
			})),
			goals: []string{"all"},
			wantTargetSetsNeedingBuild: [][]string{{"compile"}, {"all"}},
		},
	}

	for label, test := range tests {
		conf := &Config{FS: test.fs}
		mk := conf.NewMaker(test.mf, test.goals...)
		if test.afterMake != nil {
			if err := mk.Run(); err != nil {
				t.Fatal(err)
			}
			if err := test.afterMake(test.fs); err != nil {
				t.Fatal(err)
			}
		}
		targetSets, err := mk.TargetSetsNeedingBuild()
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: TargetsNeedingBuild(%q): error: %s", label, test.goals, err)
				continue
			} else {
				t.Errorf("%s: TargetsNeedingBuild(%q): error: got %q, want %q", label, test.goals, err, test.wantErr)
				continue
			}
		}

		// sort so that test is deterministic
		for _, ts := range targetSets {
			sort.Strings(ts)
		}
		if !reflect.DeepEqual(targetSets, test.wantTargetSetsNeedingBuild) {
			t.Errorf("%s: got targetSets needing build %v, want %v", label, targetSets, test.wantTargetSetsNeedingBuild)
		}
	}
}
