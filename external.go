package makex

import (
	"io/ioutil"
	"os"
	"os/exec"
)

func ExternalMake(dir string, makefile []byte, args []string) error {
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
