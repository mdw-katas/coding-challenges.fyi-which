package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var Version = "dev"

func main() {
	log.SetFlags(log.Lshortfile)
	var (
		findAll bool
		silent  bool
	)
	programName := fmt.Sprintf("%s @ %s", filepath.Base(os.Args[0]), Version)
	flags := flag.NewFlagSet(programName, flag.ExitOnError)
	flags.BoolVar(&findAll, "a", false,
		"List all instances of executables found "+
			"(instead of just the first one of each).")
	flags.BoolVar(&silent, "s", false,
		"No output, just return 0 if all of the executables "+
			"are found, or 1 if some were not found.")
	flags.Usage = func() {
		output := flags.Output()
		_, _ = fmt.Fprintf(output, "Usage of %s:\n", flags.Name())
		_, _ = fmt.Fprintf(output, "%s [-as] program ...\n", filepath.Base(os.Args[0]))
		_, _ = fmt.Fprintf(output, "which – locate a program file in the user's path")
		_, _ = fmt.Fprintln(output, "See `man which` for the builtin which program's usage.")
		flags.PrintDefaults()
	}
	_ = flags.Parse(os.Args[1:])
	arguments := flags.Args()
	if len(arguments) == 0 {
		log.Fatal("At least one executable is required.")
	}

	programs := make(map[string]struct{}, len(arguments))
	for _, arg := range arguments {
		programs[arg] = struct{}{}
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
			_, contains := programs[name]
			if contains {
				if !silent {
					fmt.Println(fullPath)
				}
				if !findAll {
					delete(programs, name)
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
