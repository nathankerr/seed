package main

// standard packages
import (
	"flag"
	"fmt"
	"github.com/nathankerr/seed"
	"github.com/nathankerr/seed/examples"
	"github.com/nathankerr/seed/executor"
	"github.com/nathankerr/seed/executor/bud"
	"github.com/nathankerr/seed/executor/monitor"
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
	var outputdir = flag.String("o", "build",
		"directory name to create and output the bud source")
	var from_format = flag.String("f", "seed",
		"format to load (seed, json)")
	var full = flag.Bool("full", false, "when true, seed input is not limited to the subset")
	var to_format = flag.String("t", "go",
		"formats to write separated by spaces (bloom, dot, go, json, seed)")
	var transformations = flag.String("transformations", "network replicate",
		"transformations to perform, separated by spaces")
	var execute = flag.Bool("execute", false, "execute the seed")
	var timeout = flag.String("timeout", "", "how long to run; if 0, run forever")
	var sleep = flag.String("sleep", "", "how long to sleep each timestep")
	var address = flag.String("address", ":3000", "address the bud communicator uses")
	var monitorAddress = flag.String("monitor", "", "address to access the debugger (http), empty means the debugger doesn't run")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s ", os.Args[0])
		fmt.Fprintf(os.Stderr, "[options] [input filename]\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	info("Load")
	filename := flag.Arg(0)
	filename = filepath.Clean(filename)
	service, name := load(filename, *from_format, *full)

	info("Transform")
	for _, transformation := range strings.Fields(*transformations) {
		service = transform(service, transformation)
	}

	info("Write")
	write(service, service.Name, *to_format, *outputdir)

	if *execute {
		info("Execute")
		start(service, name, *sleep, *timeout, *address, *monitorAddress)
	}
}

func load(filename, format string, full bool) (*seed.Seed, string) {
	_, name := filepath.Split(filename)
	name = name[:len(name)-len(filepath.Ext(name))]

	source, err := ioutil.ReadFile(filename)
	if err != nil {
		fatal(err)
	}

	var service *seed.Seed
	switch format {	
	case "seed":
		service, err = seed.FromSeed(filename, source, !full)
	case "json":
		service, err = seed.FromJson(filename, source)
	default:
		fatal("Loading from", format, "format not supported.\n")
	}
	if err != nil {
		fatal(err)
	}

	service.Name = name

	err = service.Validate()
	if err != nil {
		fatal(err)
	}

	return service, name
}

func write(service *seed.Seed, name string, formats string, outputdir string) {
	outputdir = filepath.Clean(outputdir)
	err := os.MkdirAll(outputdir, 0755)
	if err != nil {
		fatal(err)
	}

	for _, format := range strings.Fields(formats) {
		var extension string
		var writer func(service *seed.Seed, name string) ([]byte, error)
		switch format {
		case "bloom":
			extension = "rb"
			writer = seed.ToBloom
		case "dot":
			extension = "dot"
			writer = seed.ToDot
		case "go":
			extension = "go"
			writer = seed.ToGo
		case "json":
			extension = "json"
			writer = seed.ToJson
		case "seed":
			extension = "seed"
			writer = seed.ToSeed
		case "latex":
			extension = "latex"
			writer = seed.ToLaTeX
		default:
			fatal("Writing to", format, "format not supported.\n")
		}

		filename := filepath.Join(outputdir,
			strings.ToLower(name)+"."+extension)
		out, err := os.Create(filename)
		if err != nil {
			fatal(err)
		}

		marshalled, err := writer(service, name)
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

func transform(service *seed.Seed, transformation string) *seed.Seed {
	var transform func(service *seed.Seed) (*seed.Seed, error)
	switch transformation {
	case "network":
		transform = examples.Add_network_interface
	case "replicate":
		transform = examples.Add_replicated_tables
	default:
		fatal(transformation, "not supported.\n")
	}

	transformed, err := transform(service)
	if err != nil {
		fatal(transformation, "error:", err)
	}

	return transformed
}

func start(service *seed.Seed, name, sleep, timeout, address, monitorAddress string) {
	info("Starting", name)

	var err error
	var sleepDuration time.Duration
	if sleep != "" {
		sleepDuration, err = time.ParseDuration(sleep)
		if err != nil {
			fatal(err)
		}
	}

	var timeoutDuration time.Duration
	if timeout != "" {
		timeoutDuration, err = time.ParseDuration(timeout)
		if err != nil {
			fatal(err)
		}
	}

	channels := executor.Execute(service, timeoutDuration, sleepDuration, address, monitorAddress)
	go monitor.StartMonitor(monitorAddress, channels.Monitor, service)
	bud.BudCommunicator(service, channels, address)
}
