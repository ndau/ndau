package version

import (
	"errors"
	"fmt"
	"os"
)

// This is set externally during the build process.
// See docker/ndaunode-build.docker
var version string

// Emit version information and quit
func Emit() {
	v, e := Get()
	if e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
	fmt.Println(v)
	os.Exit(0)
}

// Get the current version or err if unset
func Get() (string, error) {
	var err error
	if len(version) == 0 {
		err = errors.New("bad build: VERSION not set")
	}
	return version, err
}
