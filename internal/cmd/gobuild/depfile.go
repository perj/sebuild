// Copyright 2019 Schibsted

package gobuild

import (
	"bytes"
	"fmt"
	"os"
)

func writeDepfile(depf *os.File) error {
	defer depf.Close()

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s: ", outpath)
	err := runWithBuildFlagsAndPkg(&buf, "go", "list", "-deps", "-f", `{{$dir:=.Dir}}{{range .GoFiles}}{{$dir}}/{{.}} {{end}}{{range .CgoFiles}}{{$dir}}/{{.}} {{end}}{{range .HFiles}}{{$dir}}/{{.}} {{end}}{{range .CFiles}}{{$dir}}/{{.}} {{end}}{{range .TestGoFiles}}{{$dir}}/{{.}} {{end}}`)
	if err != nil {
		return err
	}
	deps := buf.Bytes()
	for i, b := range deps {
		if b == '\n' {
			deps[i] = ' '
		}
	}
	deps = append(deps, '\n')
	_, err = depf.Write(deps)
	return err
}
