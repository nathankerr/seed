package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// flow:
// - load seeds
// - add network interface
// - write bud to ruby
func main() {
	log.SetFlags(0) // turn off logger prefix

	var outputdir = flag.String("o", "bud", "directory name to create and output the bud source")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s ", os.Args[0])
		fmt.Fprintf(os.Stderr, "[options] [input files]\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// load seeds
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	seeds, err := loadSeeds(flag.Args())
	if err != nil {
		log.Fatalln(err)
	}

	// Add network interface
	buds := make(budCollection)
	for sname, seed := range seeds {
		clusters := getClusters(sname, seed)

		for _, cluster := range clusters {
			buds = add_clients(buds, cluster, seed, sname)
		}
	}

	// write bud to ruby
	err = buds.toRuby(*outputdir)
	if err != nil {
		panic(err)
	}

}
