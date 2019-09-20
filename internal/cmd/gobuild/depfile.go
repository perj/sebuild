// Copyright 2019 Schibsted

package gobuild

import (
	"bufio"
	"bytes"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"runtime"
)

func writeDepfile(depf *os.File) error {
	defer depf.Close()

	var buf bytes.Buffer
	err := runWithBuildFlagsAndPkg(&buf, "go", "list", "-deps", "-f", `{{$dir:=.Dir}}{{range .GoFiles}}{{$dir}}/{{.}} {{end}}{{range .CgoFiles}}{{$dir}}/{{.}} {{end}}{{range .HFiles}}{{$dir}}/{{.}} {{end}}{{range .CFiles}}{{$dir}}/{{.}} {{end}}{{range .TestGoFiles}}{{$dir}}/{{.}} {{end}}`)
	if err != nil {
		return err
	}

	// Ignore files in GOROOT and in modules. These should not normally
	// change, except on Go version upgrades, but I think it's a fair trade
	// with dependency size to have to manually rebuild after upgrading Go.
	goroot := []byte(runtime.GOROOT() + "/")
	gomodroot := []byte(filepath.Join(build.Default.GOPATH, "pkg/mod") + "/")

	w := bufio.NewWriter(depf)
	fmt.Fprintf(w, "%s:", outpath)
	s := bufio.NewScanner(&buf)
	s.Split(bufio.ScanWords)
	for s.Scan() {
		dep := bytes.TrimSpace(s.Bytes())
		if bytes.HasPrefix(dep, goroot) || bytes.HasPrefix(dep, gomodroot) {
			continue
		}
		w.WriteRune(' ')
		w.Write(dep)
	}
	w.WriteRune('\n')
	return w.Flush()
}
