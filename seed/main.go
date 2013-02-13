package main

// standard packages
import (
	"flag"
	"fmt"
	service "github.com/nathankerr/seed"
	"github.com/nathankerr/seed/examples"
	"github.com/nathankerr/seed/executor"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// flow:
// - load seeds
// - add network interface
// - write bud to ruby
func main() {
	var outputdir = *flag.String("o", "build",
		"directory name to create and output the bud source")
	var from_format = flag.String("f", "seed",
		"format to load (seed)")
	var to_format = flag.String("t", "go",
		"formats to write separated by spaces (bloom, dot, go, json, service)")
	var transformations = flag.String("transformations", "network replicate",
		"transformations to perform, separated by spaces")
	var execute = flag.Bool("execute", false, "execute the service")
	var timeout = flag.String("timeout", "", "how long to run; if 0, run forever")
	var sleep = flag.String("sleep", "", "how long to sleep each timestep")
	var address = flag.String("address", "127.0.0.1:3000", "address the bud communicator uses")
	var monitor = flag.String("monitor", "", "address to access the debugger (http), empty means the debugger doesn't run")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s ", os.Args[0])
		fmt.Fprintf(os.Stderr, "[options] [input files]\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	info("Load Seeds")
	seeds := make(map[string]*service.Service)

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

		var seed *service.Service
		switch *from_format {
		case "seed":
			seed = service.Parse(filename, string(seedSource))
		default:
			fatal("Loading from", *from_format, "format not supported.\n")
		}

		err = seed.Validate()
		if err != nil {
			fatal(err)
		}

		seeds[name] = seed
	}

	info("Transform Seeds")
	for _, transformation := range strings.Fields(*transformations) {
		transformed := make(map[string]*service.Service)
		for sname, seed := range seeds {
			var transform func(name string, seed *service.Service, seeds map[string]*service.Service) map[string]*service.Service
			switch transformation {
			case "network":
				transform = examples.Add_network_interface
			case "replicate":
				transform = examples.Add_replicated_tables
			default:
				fatal(transformation, "not supported.\n")
			}

			transformed = transform(sname, seed, transformed)
		}
		seeds = transformed
	}

	info("Write Seeds")
	outputdir = filepath.Clean(outputdir)
	err := os.MkdirAll(outputdir, 0755)
	if err != nil {
		fatal(err)
	}

	for name, seed := range seeds {
		for _, format := range strings.Fields(*to_format) {
			var extension string
			var writer func(seed *service.Service, name string) ([]byte, error)
			switch format {
			case "bloom":
				extension = "rb"
				writer = service.ToBloom
			case "dot":
				extension = "dot"
				writer = service.ToDot
			case "go":
				extension = "go"
				writer = service.ToGo
			case "json":
				extension = "json"
				writer = service.ToJson
			case "service":
				extension = "service"
				writer = service.ToModel
			default:
				fatal("Writing to", format, "format not supported.\n")
			}

			filename := filepath.Join(outputdir,
				strings.ToLower(name)+"."+extension)
			out, err := os.Create(filename)
			if err != nil {
				fatal(err)
			}

			marshalled, err := writer(seed, name)
			if err != nil {
				fatal("Error while converting to", format, ":", err)
			}
			_, err = out.Write(marshalled)
			if err != nil {
				fatal(err)
			}

			out.Close()
		}
	}

	if *execute {
		info("Execute")
		for name, seed := range seeds {
			info("Starting", name)

			var sleepDuration time.Duration
			if *sleep != "" {
				sleepDuration, err = time.ParseDuration(*sleep)
				if err != nil {
					fatal(err)
				}
			}

			var timeoutDuration time.Duration
			if *timeout != "" {
				timeoutDuration, err = time.ParseDuration(*timeout)
				if err != nil {
					fatal(err)
				}
			}

			executor.Execute(seed, timeoutDuration, sleepDuration, *address, *monitor)
			break
		}
	}
}
