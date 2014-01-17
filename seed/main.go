package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/nathankerr/seed"
	"github.com/nathankerr/seed/examples"
	"github.com/nathankerr/seed/executor"
	"github.com/nathankerr/seed/executor/bud"
	"github.com/nathankerr/seed/executor/monitor"
	"github.com/nathankerr/seed/executor/tracer"
	"github.com/nathankerr/seed/executor/wsjson"
	"github.com/nathankerr/seed/host/bloom"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)

	var outputdir = flag.String("o", "build",
		"directory name to create and output the bud source")
	var from_format = flag.String("f", "seed",
		"format to load (seed, json)")
	var full = flag.Bool("full", false,
		"when true, seed input is not limited to the subset")
	var to_format = flag.String("t", "go",
		"formats to write separated by spaces (bloom, dot, go, json, seed)")
	var transformations = flag.String("transformations", "network replicate",
		"transformations to perform, separated by spaces")
	var execute = flag.Bool("execute", false,
		"execute the seed")
	var sleep = flag.String("sleep", "",
		"how long to sleep each timestep")
	var address = flag.String("address", ":3000",
		"address the communicator uses")
	var communicator = flag.String("communicator", "wsjson",
		"which communicator to use (bud, wsjson")
	var monitorAddress = flag.String("monitor", "",
		"address the monitor uses; empty means it will not run")
	var traceFilename = flag.String("trace", "",
		"filename to dump a trace to; empty means it will not run")
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

	log.Println("Load")
	filename := flag.Arg(0)
	filename = filepath.Clean(filename)
	service, err := load(filename, *from_format, *full)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Transform")
	for _, transformation := range strings.Fields(*transformations) {
		service, err = transform(service, transformation)
		if err != nil {
			log.Fatalln(err)
		}
	}

	log.Println("Write")
	write(service, service.Name, *to_format, *outputdir)

	if *execute {
		log.Println("Executing")
		err = start(service, *sleep, *address, *monitorAddress, *traceFilename, *communicator)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func load(filename, format string, full bool) (*seed.Seed, error) {
	_, name := filepath.Split(filename)
	name = name[:len(name)-len(filepath.Ext(name))]

	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var service *seed.Seed
	switch format {
	case "seed":
		service, err = seed.FromSeed(filename, source)
	case "json":
		service, err = seed.FromJson(filename, source)
	default:
		return nil, errors.New(fmt.Sprint("Loading from", format, "format not supported."))
	}
	if err != nil {
		return nil, err
	}

	service.Name = name

	err = service.Validate()
	if err != nil {
		return nil, err
	}

	if !full {
		err = service.InSubset()
		if err != nil {
			return nil, err
		}
	}

	return service, nil
}

func write(service *seed.Seed, name string, formats string, outputdir string) {
	outputdir = filepath.Clean(outputdir)
	err := os.MkdirAll(outputdir, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	for _, format := range strings.Fields(formats) {
		var extension string
		var writer func(service *seed.Seed, name string) ([]byte, error)
		switch format {
		case "bloom":
			extension = "rb"
			writer = bloom.ToBloom
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
		default:
			log.Fatalln("Writing to", format, "format not supported.\n")
		}

		filename := filepath.Join(outputdir,
			strings.ToLower(name)+"."+extension)
		out, err := os.Create(filename)
		if err != nil {
			log.Fatalln(err)
		}

		marshalled, err := writer(service, name)
		if err != nil {
			log.Fatalln("Error while converting to", format, ":", err)
		}
		_, err = out.Write(marshalled)
		if err != nil {
			log.Fatalln(err)
		}

		out.Close()
	}
}

func transform(service *seed.Seed, transformation string) (*seed.Seed, error) {
	var transform func(service *seed.Seed) (*seed.Seed, error)
	switch transformation {
	case "network":
		transform = examples.Add_network_interface
	case "network-graph":
		transform = examples.Add_network_interface_graph
	case "replicate":
		transform = examples.Add_replicated_tables
	default:
		return nil, errors.New(transformation + " not supported.")
	}

	transformed, err := transform(service)
	if err != nil {
		return nil, errors.New(fmt.Sprint(transformation, ": ", err))
	}

	return transformed, nil
}

func start(service *seed.Seed, sleep, address, monitorAddress, traceFilename, communicator string) error {
	var err error
	var sleepDuration time.Duration
	if sleep != "" {
		sleepDuration, err = time.ParseDuration(sleep)
		if err != nil {
			return err
		}
	}

	useMonitor := false
	if (monitorAddress != "") || (traceFilename != "") {
		useMonitor = true
	}

	if (monitorAddress != "") && (traceFilename != "") {
		log.Fatalln("cannot use both the web-based monitoring and tracing")
	}

	channels := executor.Execute(service, sleepDuration, address, useMonitor)

	if monitorAddress != "" {
		go monitor.StartMonitor(monitorAddress, channels, service)
	} else if traceFilename != "" {
		go tracer.StartTracer(traceFilename, channels.Monitor)
	}

	switch communicator {
	case "bud":
		bud.BudCommunicator(service, channels, address)
	case "wsjson":
		wsjson.Communicator(service, channels, address)
	default:
		return errors.New("Unknown communicator: " + communicator)
	}

	return nil
}
