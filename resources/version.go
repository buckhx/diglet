// package diglet/resources contains string resources about how to run diglet.
// Files generated from scripts/include.go will end up here as well as deployment utils like version.
package resources

import (
	"bytes"
	"os/exec"
	"strings"
)

func Version() string {
	cmd := exec.Command("git", "describe", "--always")
	var ver bytes.Buffer
	cmd.Stdout = &ver
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(ver.String())
	//TODO add a '{+n}' to version if git diff --numstat isn't empty
}
