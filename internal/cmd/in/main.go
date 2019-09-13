// Copyright 2019 Schibsted

package in

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func Main(args ...string) {
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s -tool in <variable-file> <in> <out>\n", os.Args[0])
		os.Exit(1)
	}
	variables := readVariables(args[0])
	in := args[1]
	out := args[2]

	bs := variables["BUILD_STAGE"]
	if bs == "" {
		fmt.Fprintf(os.Stderr, "No BUILD_STAGE variable in %s.\n", args[0])
		os.Exit(1)
	}

	oldout, err := ioutil.ReadFile(out)
	if err != nil {
		oldout = nil
	}

	var buf bytes.Buffer
	f, err := os.Open(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open in file %s: %s\n", in, err)
		os.Exit(1)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	startRE := regexp.MustCompile("%[A-Z]+_START%\n")
	endRE := regexp.MustCompile("%[A-Z]+_END%\n")
	skipStage := ""
	lineno := 0
	for {
		line, err := r.ReadBytes('\n')
		lineno++
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read in file %s: %s\n", in, err)
			os.Exit(1)
		}
		if start := startRE.FindIndex(line); start != nil {
			stage := string(line[start[0]+1 : start[1]-len("%_START%")])
			if stage != bs && stage != "ANY" {
				skipStage = stage
			}
			line = append(line[:start[0]], line[start[1]:]...)
		}
		if end := endRE.FindIndex(line); end != nil {
			stage := string(line[end[0]+1 : end[1]-len("%_END%")])
			if stage == skipStage {
				skipStage = ""
			}
			line = append(line[:end[0]], line[end[1]:]...)
		}
		if skipStage != "" {
			continue
		}
		percs := bytes.Split(line, []byte("%"))
		for i := 0; i < len(percs)-1; i += 2 {
			buf.Write(percs[i])
			if i == len(percs)-1 {
				fmt.Fprintf(os.Stderr, "Error in %s:%d: Unmatched %.\n", in, lineno)
				os.Exit(1)
			}
			key := string(percs[i+1])
			if key == "" {
				buf.WriteByte('%')
				continue
			}
			v, ok := variables[key]
			if !ok {
				fmt.Fprintf(os.Stderr, "Error in %s:%d: undefined variable >%s< -> %s\n", in, lineno, key, line)
				os.Exit(1)
			}
			buf.WriteString(v)
		}
		buf.Write(percs[len(percs)-1])
	}

	newout := buf.Bytes()
	if oldout != nil && bytes.Equal(oldout, newout) {
		return
	}
	err = ioutil.WriteFile(out, newout, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to output %s: %s\n", out, err)
		os.Exit(1)
	}
}

func readVariables(file string) map[string]string {
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open variables file %s: %s\n", file, err)
		os.Exit(1)
	}
	defer f.Close()

	ret := make(map[string]string)
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read variables file %s: %s\n", file, err)
			os.Exit(1)
		}
		vals := strings.SplitN(strings.TrimSpace(line), "=", 2)
		if len(vals) == 2 {
			ret[vals[0]] = vals[1]
		}
	}
	return ret
}
