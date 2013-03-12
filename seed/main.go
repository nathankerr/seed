package main

// standard packages
import (
	"flag"
	"fmt"
	service2 "github.com/nathankerr/seed"
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
	var outputdir = *flag.String("o", "build",
		"directory name to create and output the bud source")
	var from_format = flag.String("f", "seed",
		"format to load (seed, json)")
	var full = flag.Bool("full", false, "when true, seed input is not limited to the subset")
	var to_format = flag.String("t", "go",
		"formats to write separated by spaces (bloom, dot, go, json, seed)")
	var transformations = flag.String("transformations", "network replicate",
		"transformations to perform, separated by spaces")
	var execute = flag.Bool("execute", false, "execute the service2")
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
	seed, name := load(filename, *from_format, *full)

	info("Transform")
	for _, transformation := range strings.Fields(*transformations) {
		seed = transform(seed, transformation)
	}

	info("Write")
	write(seed, name, *to_format, outputdir)

	if *execute {
		info("Execute")
		start(seed, name, *sleep, *timeout, *address, *monitorAddress)
	}
}

func load(filename, format string, full bool) (*service2.Seed, string) {
	_, name := filepath.Split(filename)
	name = name[:len(name)-len(filepath.Ext(name))]

	source, err := ioutil.ReadFile(filename)
	if err != nil {
		fatal(err)
	}

	var service *service2.Seed
	switch format {
	case "seed":
		service, err = service2.FromSeed(filename, source, !full)
	case "json":
		service, err = service2.FromJson(filename, source)
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

func write(seed *service2.Seed, name string, formats string, outputdir string) {
	outputdir = filepath.Clean(outputdir)
	err := os.MkdirAll(outputdir, 0755)
	if err != nil {
		fatal(err)
	}

	for _, format := range strings.Fields(formats) {
		var extension string
		var writer func(seed *service2.Seed, name string) ([]byte, error)
		switch format {
		case "bloom":
			extension = "rb"
			writer = service2.ToBloom
		case "dot":
			extension = "dot"
			writer = service2.ToDot
		case "go":
			extension = "go"
			writer = service2.ToGo
		case "json":
			extension = "json"
			writer = service2.ToJson
		case "seed":
			extension = "seed"
			writer = service2.ToSeed
		case "latex":
			extension = "latex"
			writer = service2.ToLaTeX
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

func transform(seed *service2.Seed, transformation string) *service2.Seed {
	var transform func(seed *service2.Seed) (*service2.Seed, error)
	switch transformation {
	case "network":
		transform = examples.Add_network_interface
	case "replicate":
		transform = examples.Add_replicated_tables
	default:
		fatal(transformation, "not supported.\n")
	}

	transformed, err := transform(seed)
	if err != nil {
		fatal(transformation, "error:", err)
	}

	return transformed
}

func start(seed *service2.Seed, name, sleep, timeout, address, monitorAddress string) {
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

	channels := executor.Execute(seed, timeoutDuration, sleepDuration, address, monitorAddress)
	go monitor.StartMonitor(monitorAddress, channels.Monitor, seed)
	bud.BudCommunicator(seed, channels, address)
}
