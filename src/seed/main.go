package main

// standard packages
import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// others
import (
	"service"
	net_transform "network"
	"replication"
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
	var json = flag.Bool("json", false,
		"produce json versions of the services")
	var model = flag.Bool("model", false,
		"produce seed like versions of the services")
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

	seeds := make(map[string]*service.Service)
	var transformed map[string]*service.Service

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

		seed := service.Parse(filename, string(seedSource))
		seeds[name] = seed
	}

	if *network {
		info("Add Network Interface")
		transformed := make(map[string]*service.Service)
		for sname, seed := range seeds {
			transformed = net_transform.Add_network_interface(sname, seed, transformed)
		}
		seeds = transformed
	}

	if *replicate {
		info("Add Replicated Tables")
		transformed = make(map[string]*service.Service)
		for sname, seed := range seeds {
			transformed = replication.Add_replicated_tables(sname, seed, transformed)
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

		ruby := bud.ToRuby(name)
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

			ruby := bud.ToDot(name)
			_, err = out.Write([]byte(ruby))
			if err != nil {
				fatal(err)
			}

			out.Close()
		}
	}

	if *json {
		info("Write json")

		for name, bud := range seeds {
			filename := filepath.Join(outputdir, strings.ToLower(name)+".json")
			out, err := os.Create(filename)
			if err != nil {
				fatal(err)
			}

			ruby := bud.ToJson(name)
			_, err = out.Write([]byte(ruby))
			if err != nil {
				fatal(err)
			}

			out.Close()
		}
	}

	if *model {
		info("Write model")

		for name, bud := range seeds {
			filename := filepath.Join(outputdir, strings.ToLower(name)+".model")
			out, err := os.Create(filename)
			if err != nil {
				fatal(err)
			}

			model := bud.ToModel(name)
			_, err = out.Write([]byte(model))
			if err != nil {
				fatal(err)
			}

			out.Close()
		}
	}
}
