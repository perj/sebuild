package touch

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"
)

var (
	flagset = flag.NewFlagSet("touch", flag.ExitOnError)
	offset  = flagset.Duration("offset", 0, "Offset to add to the time set.")
)

func Main(args ...string) {
	flagset.Usage = func() {
		fmt.Fprintf(flagset.Output(), "Usage: %s -tool touch [options] <files...>\n", os.Args[0])
		flagset.PrintDefaults()
	}
	flagset.Parse(args)
	if flagset.NArg() == 0 {
		flagset.Usage()
		os.Exit(1)
	}
	ts := time.Now().Add(*offset)

	ret := 0
	for _, name := range flagset.Args() {
		err := os.Chtimes(name, ts, ts)
		if err != nil {
			perr := err.(*os.PathError)
			if perr.Err == syscall.ENOENT {
				var f *os.File
				f, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0666)
				if f != nil {
					f.Close()
				}
				if err == nil {
					err = os.Chtimes(name, ts, ts)
				}
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
			ret = 1
		}
	}
	os.Exit(ret)
}
