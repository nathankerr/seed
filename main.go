package main

import (
	"flag"
	"fmt"
	"os"
)

// flow:
// - load seeds
// - apply seed-> seed transforms
// - apply seed -> bud transforms
// - apply bud -> bud transforms
// - write bud to ruby
func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("USAGE: seed [seed files]")
		os.Exit(1)
	}

	// load seeds
	seeds, err := loadSeeds(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	// apply seed-> seed transforms
	seeds = applySeedTransforms(seeds, nil)

	// apply seed -> bud transforms
	buds := applySeedToBudTransforms(seeds, nil)

	// apply bud -> bud transforms
	buds = applyBudTransforms(buds, nil)

	// write bud to ruby
	err = buds.toRuby("buds")
	if err != nil {
		panic(err)
	}

	for name, seed := range seeds {
		fmt.Println("# ", name)
		fmt.Println(seed)
	}
}
