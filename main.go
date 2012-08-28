package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	s2s := []seedTransform{splitSeeds}
	seeds = applySeedTransforms(seeds, s2s)

	for name, seed := range(seeds) {
		fmt.Printf("### %s ###\n%s\n\n", name, seed)
	}

	// apply seed -> bud transforms
	s2b := []seedToBudTransform{generateServer, generateClient}
	buds := applySeedToBudTransforms(seeds, s2b)

	// apply bud -> bud transforms
	buds = applyBudTransforms(buds, nil)

	// write bud to ruby
	err = buds.toRuby(*outputdir)
	if err != nil {
		panic(err)
	}

}
