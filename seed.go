package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type seed struct {
	collections map[string]*table
	rules       []*rule
}

func newSeed() *seed {
	return &seed{
		collections: make(map[string]*table),
	}
}

func (s *seed) String() string {
	str := "collections:"
	for k, v := range s.collections {
		str = fmt.Sprintf("%s\n\t%s %s %s\t(%s)", str, v.typ, k, v, v.source)
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

func (seeds seedCollection) String() string {
	str := ""
	for name, seed := range seeds {
		str = fmt.Sprintf("%s\n### seed: %s ###\n%s\n", str, name, seed)
	}
	return str
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

type seedCollectionType int

const (
	seedInput seedCollectionType = iota
	seedOutput
	seedTable
)

var seedCollectionTypeNames = map[seedCollectionType]string{
	seedInput:  "input",
	seedOutput: "output",
	seedTable:  "table",
}

func (t seedCollectionType) String() string {
	str, ok := seedCollectionTypeNames[t]
	if !ok {
		panic("unknown seed collection type")
	}
	return str
}

type table struct {
	key     []string
	columns []string
	source  source
	typ     seedCollectionType
}

func newTable(typ seedCollectionType) *table {
	return &table{}
}

func (t *table) String() string {
	return fmt.Sprint(t.key, "=>", t.columns)
}
