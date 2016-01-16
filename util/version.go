package util

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
