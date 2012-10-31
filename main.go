package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// flow:
// - load seeds
// - add network interface
// - write bud to ruby
func main() {
	log.SetFlags(0) // turn off logger prefix

	var outputdir = flag.String("o", "bud",
		"directory name to create and output the bud source")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s ", os.Args[0])
		fmt.Fprintf(os.Stderr, "[options] [input files]\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	info("Load Seeds")
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	seeds := make(map[string]*service)

	for _, filename := range flag.Args() {
		filename = filepath.Clean(filename)

		_, name := filepath.Split(filename)
		name = name[:len(name)-len(filepath.Ext(name))]
		if _, ok := seeds[name]; ok {
			fatal("Seed", name, "already exists")
		}

		seedSource, err := ioutil.ReadFile(filename)
		if err != nil {
			fatal(err)
		}

		seed := parse(filename, string(seedSource))
		seeds[name] = seed
	}

	// Add network interface
	buds := make(map[string]*service)
	for sname, seed := range seeds {
		groups := getGroups(sname, seed)

		for _, group := range groups {
			switch group.typ() {
			case "000", "010", "0n0", "100", "n00":
				panic("these clusters should not exist in seed")
			case "011", "01n", "0n1", "0nn":
				// fatal(group.typ(), "not supported in", sname)
				// works with assumption that first column is address
				buds = add_clients(buds, group, seed, sname)
			case "101", "10n", "n01", "n0n", "001", "00n", "110", "111",
				"11n", "1n0", "1n1", "1nn", "n10", "n11", "n1n", "nn0",
				"nn1", "nnn":
				buds = add_clients(buds, group, seed, sname)
			default:
				panic("should not get here")
			}
		}
	}

	// write bud to ruby
	for name, seed := range seeds {
		err := seed.toRuby(name, *outputdir)
		if err != nil {
			panic(err)
		}
	}
}
