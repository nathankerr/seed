package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// flow:
// - load seeds
// - add network interface
// - write bud to ruby
func main() {
	var outputdir = *flag.String("o", "bud",
		"directory name to create and output the bud source")
	var network = flag.Bool("network", true,
		"add network interface")
	var replicate = flag.Bool("replicate", true,
		"replicate tables")
	var dot = flag.Bool("dot", false,
		"also produce dot (graphviz) files)")
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
	var transformed map[string]*service

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

	if *network {
		info("Add Network Interface")
		transformed := make(map[string]*service)
		for sname, seed := range seeds {
			groups := getGroups(sname, seed)

			for _, group := range groups {
				switch group.typ() {
				case "000", "010", "0n0", "100", "n00": // not possible
					panic(group.typ())
				case "011", "01n", "0n1", "0nn": // not input driven (not handled)
					panic(group.typ())
				case "001", "00n", "101", "10n", "n01", "n0n", // passthrough
					"110", "111", "11n", // single output
					"1n0", "1n1", "1nn", // multiple output
					"n10", "n11", "n1n", "nn0", "nn1", "nnn": // multiple input
					transformed = add_network_interface(transformed, group, seed, sname)
				default:
					// shouldn't get here
					panic(group.typ())
				}
			}
		}
		seeds = transformed
	}

	if *replicate {
		info("Add Replicated Tables")
		transformed = make(map[string]*service)
		for sname, seed := range seeds {
			transformed = add_replicated_tables(sname, seed, transformed)
		}
		seeds = transformed
	}

	info("Write Ruby")
	outputdir = filepath.Clean(outputdir)
	err := os.MkdirAll(outputdir, 0755)
	if err != nil {
		fatal(err)
	}

	for name, bud := range seeds {
		filename := filepath.Join(outputdir, strings.ToLower(name)+".rb")
		out, err := os.Create(filename)
		if err != nil {
			fatal(err)
		}

		ruby := bud.toRuby(name)
		_, err = out.Write([]byte(ruby))
		if err != nil {
			fatal(err)
		}

		out.Close()
	}

	if *dot {
		info("Write Dot")

		for name, bud := range seeds {
			filename := filepath.Join(outputdir, strings.ToLower(name)+".dot")
			out, err := os.Create(filename)
			if err != nil {
				fatal(err)
			}

			ruby := bud.toDot(name)
			_, err = out.Write([]byte(ruby))
			if err != nil {
				fatal(err)
			}

			out.Close()
		}
	}
}
