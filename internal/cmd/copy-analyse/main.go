package copy_analyse

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func Main(args ...string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s -tool copy-analyse [-finalize] <destdir> <srcdir...>\n", os.Args[0])
		os.Exit(1)
	}

	final := args[0] == "-finalize"
	if final {
		args = args[1:]
	}

	out := args[0]
	os.RemoveAll(out)
	if err := os.Mkdir(out, 0777); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory %s: %s\n", out, err)
		os.Exit(1)
	}

	ecode := 0
	for _, src := range args[1:] {
		if err := copyAnalyse(out, src); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to copy %s to %s: %s\n", src, out, err)
			ecode = 1
		}
	}
	if ecode == 0 && final {
		if err := finalize(out); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to finalize %s: %s\n", out, err)
			ecode = 1
		}
	}
	os.Exit(ecode)
}

func copyAnalyse(out, in string) error {
	return filepath.Walk(in, func(path string, info os.FileInfo, err error) error {
		if err != nil || filepath.Ext(path) != ".html" {
			return err
		}
		return copyFile(out, path)
	})
}

func copyFile(outdir, in string) error {
	data, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}
	outname := filepath.Base(in)
	if strings.HasPrefix(outname, "report-") {
		var ok bool
		outname, ok = checkReports(outname, data)
		if !ok {
			return nil
		}
	}
	return ioutil.WriteFile(filepath.Join(outdir, outname), data, 0666)
}

var falsePositives = []string{
	"TAILQ_REMOVE",
	"RB_GENERATE_STATIC",
	"ALLOC_PERMANENT_ZVAL",
}

func checkReports(name string, data []byte) (string, bool) {
	s := bufio.NewScanner(bytes.NewReader(data))
	var search string
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "<title>") {
			newname := strings.TrimSuffix(strings.TrimPrefix(line, "<title>"), "</title>")
			name = strings.ReplaceAll(name, "report", newname)
			name = strings.ReplaceAll(name, "/", "-")
		}
		var buglnum int
		n, _ := fmt.Sscanf(line, "<!-- BUGLINE %d -->", &buglnum)
		if n == 1 {
			search = fmt.Sprintf(`id="LN%d"`, buglnum)
		}
		if search == "" {
			continue
		}
		pos := strings.Index(line, search)
		if pos == -1 {
			continue
		}
		line = line[pos:]
		for _, fp := range falsePositives {
			if strings.Contains(line, fp) {
				return name, false
			}
		}
	}
	return name, true
}

func finalize(out string) error {
	type report struct {
		Name  string
		File  string
		Title string
		Info  map[string]string
	}
	var reports []report
	err := filepath.Walk(out, func(path string, info os.FileInfo, err error) error {
		if err != nil || filepath.Ext(path) != ".html" {
			return err
		}
		inf, err := os.Open(path)
		if err != nil {
			return err
		}
		defer inf.Close()
		r := report{
			File: filepath.Base(path),
			Info: make(map[string]string),
		}
		if dash := strings.LastIndexByte(r.File, '-'); dash >= 0 {
			r.Name = r.File[:dash]
		} else {
			r.Name = r.File
		}
		s := bufio.NewScanner(inf)
		for s.Scan() {
			line := s.Text()
			if strings.HasPrefix(line, "<title>") {
				r.Title = strings.TrimSuffix(strings.TrimPrefix(line, "<title>"), "</title>")
			}
			if strings.HasPrefix(line, "<!-- BUG") {
				line = strings.TrimPrefix(line, "<!-- BUG")
				line = strings.TrimSuffix(line, "-->")
				vals := strings.SplitN(line, " ", 2)
				if len(vals) == 2 {
					v := vals[1]
					v = strings.ReplaceAll(v, "<", "&lt;")
					v = strings.ReplaceAll(v, ">", "&gt;")
					r.Info[vals[0]] = v
				}
			}
		}
		if len(r.Info) > 0 {
			reports = append(reports, r)
		}
		return s.Err()
	})
	if err != nil {
		return err
	}

	var indbuf bytes.Buffer
	ts := time.Now()
	fmt.Fprintf(&indbuf, "<html><head><title>Clang analyse report, %v</title></head>\n<body>\n", ts)
	fmt.Fprintf(&indbuf, "<p>Report generated %v.</p>\n", ts)

	sort.Slice(reports, func(i, j int) bool {
		if reports[i].Name < reports[j].Name {
			return true
		}
		return reports[i].Info["LINE"] < reports[j].Info["LINE"]
	})
	fmt.Fprintf(&indbuf, "<dl>\n")
	curfile := ""
	for _, r := range reports {
		if r.Name != curfile {
			if curfile != "" {
				fmt.Fprintf(&indbuf, "</ol></dd>\n")
			}
			fmt.Fprintf(&indbuf, "<dt>%s</dt>\n<dd><ol>\n", r.Name)
			curfile = r.Name
		}
		fmt.Fprintf(&indbuf, `<li><span title="%s: %s"><a href="%s#EndPath">line %s</a>: %s</span></li>`+"\n", r.Info["CATEGORY"], r.Info["TYPE"], r.File, r.Info["LINE"], r.Info["DESC"])
	}
	if curfile != "" {
		fmt.Fprintf(&indbuf, "</ol></dd>\n")
	}
	fmt.Fprintf(&indbuf, "</dl>\n</body>\n</html>\n")

	if len(reports) > 0 {
		fmt.Fprintf(os.Stderr, "Total of %d reports.\n", len(reports))
	}
	return ioutil.WriteFile(filepath.Join(out, "index.html"), indbuf.Bytes(), 0666)
}
