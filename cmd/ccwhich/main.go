package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var Version = "dev"

func main() {
	log.SetFlags(log.Lshortfile)
	var (
		findAll bool
		silent  bool
	)
	flags := flag.NewFlagSet(fmt.Sprintf("%s @ %s", filepath.Base(os.Args[0]), Version), flag.ExitOnError)
	flags.BoolVar(&findAll, "a", false, "List all instances of executables found (instead of just the first one of each).")
	flags.BoolVar(&silent, "s", false, "No output, just return 0 if all of the executables are found, or 1 if some were not found.")
	flags.Usage = func() {
		_, _ = fmt.Fprintf(flags.Output(), "Usage of %s:\n", flags.Name())
		_, _ = fmt.Fprintf(flags.Output(), "%s [-as] program ...\n", filepath.Base(os.Args[0]))
		_, _ = fmt.Fprintln(flags.Output(), "See man page for the builtin which program.")
		flags.PrintDefaults()
	}
	_ = flags.Parse(os.Args[1:])
	programs := flags.Args()
	if len(programs) == 0 {
		log.Fatal("At least one executable is required.")
	}

	var exitCode int
	for _, dir := range strings.Split(os.Getenv("PATH"), ":") {
		listing, err := os.ReadDir(dir)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range listing {
			if entry.IsDir() {
				continue
			}
			if !entry.Type().IsRegular() {
				continue
			}
			fullPath := filepath.Join(dir, entry.Name())
			info, _ := os.Stat(fullPath)
			permissions := info.Mode().Perm()
			if permissions&0001 == 0 {
				continue // not executable
			}
			name := entry.Name()
			if slices.Contains(programs, name) {
				if !silent {
					fmt.Println(fullPath)
				}
				if !findAll {
					programs = slices.DeleteFunc(programs, func(s string) bool { return s == name })
					if len(programs) == 0 {
						os.Exit(exitCode)
					}
				}
			} else {
				exitCode = 1
			}
		}
	}
	os.Exit(exitCode)
}
