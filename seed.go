package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type seed struct {
	inputs  tableCollection
	outputs tableCollection
	tables  tableCollection
	rules   []*rule
}

func newSeed() *seed {
	return &seed{
		inputs:  newTableCollection(),
		outputs: newTableCollection(),
		tables:  newTableCollection(),
	}
}

func (s *seed) String() string {
	str := "inputs:"
	for k, v := range s.inputs {
		str = fmt.Sprint(str, "\n\t", k, " ", v, "\t(", v.source, ")")
	}

	str += "\noutputs:"
	for k, v := range s.outputs {
		str = fmt.Sprint(str, "\n\t", k, " ", v, "\t(", v.source, ")")
	}

	str += "\ntables:"
	for k, v := range s.tables {
		str = fmt.Sprint(str, "\n\t", k, " ", v, "\t(", v.source, ")")
	}

	str += "\nrules:"
	for k, v := range s.rules {
		str = fmt.Sprint(str, "\n\t", k, "\t", v, "\t(", v.source, ")")
	}

	return str
}

type seedCollection map[string]*seed

func newSeedCollection() seedCollection {
	return make(map[string]*seed)
}

func loadSeeds(filenames []string) (seedCollection, error) {
	seeds := newSeedCollection()

	for _, filename := range filenames {
		filename = filepath.Clean(filename)
		_, name := filepath.Split(filename)
		name = name[:len(name)-len(filepath.Ext(name))]

		contents, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		seed := parse(filename, string(contents))

		seeds[name] = seed
	}

	return seeds, nil
}
