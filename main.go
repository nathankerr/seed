package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// flow:
// - load seeds
// - apply seed-> seed transforms
// - apply seed -> bud transforms
// - apply bud -> bud transforms
// - write bud to ruby
func main() {
	var outputdir = flag.String("o", "bud", "directory name to create and output the bud source")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [options] [input files]\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	log.SetFlags(0) // turn off logger prefix

	// load seeds
	seeds, err := loadSeeds(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// apply seed-> seed transforms
	// applySeedTransformations(seeds, split_seeds)

	// fmt.Print(seeds)

	// apply seed -> bud transforms
	buds := applySeedToBudTransformations(seeds,
		generate_client,
		generate_server,
	)

	// fmt.Print(buds)

	// apply bud -> bud transforms
	// buds = applyBudTransforms(buds)

	// write bud to ruby
	err = buds.toRuby(*outputdir)
	if err != nil {
		panic(err)
	}

}

func info(args ...interface{}) {
	info := ""

	pc, file, line, ok := runtime.Caller(2)
	if ok {
		basepath, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		sourcepath, err := filepath.Rel(basepath, file)
		if err != nil {
			panic(err)
		}
		info += fmt.Sprintf("%s:%d: ", sourcepath, line)

		name := path.Ext(runtime.FuncForPC(pc).Name())
		info += name[1:]
		if len(args) > 0 {
			info += ": "
		}
	}
	info += fmt.Sprintln(args...)

	log.Print(info)
}
