package core

import (
	"fmt"
	"runtime"
)

// Version to be defined at build time
var Version = "tobedefined"

// Info returns the version, the Go compiler version and the platform the binary was built for.
func Info() string {
	return fmt.Sprintf("whatismyip version %s (%s %s/%s)", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
