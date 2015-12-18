package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Reads all .txt files in the current folder
// and encodes them as strings literals in textfiles.go
func main() {
	resources := []string{}
	out, _ := os.Create("resources.go")
	out.Write([]byte("// THIS FILE IS AUTOGENERATED FROM scripts/include.go EDITS WILL BE SQUISHED\n"))
	out.Write([]byte("// IT IS MEANT FOR HOLDING CONFIGURABLE OPTIONS\n"))
	out.Write([]byte("package main\n\nconst (\n"))
	for _, resource := range resources {
		out.Write([]byte(strings.Replace(resource, ".", "_", -1) + " = `"))
		f, _ := os.Open(resource)
		io.Copy(out, f)
		out.Write([]byte("`\n"))
	}
	out.Write([]byte("Version = \"" + getVersion() + "\"\n"))
	out.Write([]byte(")\n"))
}

func getVersion() string {
	cmd := exec.Command("git", "describe", "--always")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(out.String())
}
