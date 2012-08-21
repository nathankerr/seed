package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type seed struct {
	collections  map[string]*table
	rules   []*rule
}

func newSeed() *seed {
	return &seed{
		collections:  newTableCollection(),
	}
}

func (s *seed) String() string {
	str := "collections:"
	for k, v := range s.collections {
		str = fmt.Sprintf("%s\n\t%s %s\t(%s)", str, k, v, v.source)
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

		seed, ok := parse(filename, string(contents))
		if !ok {
			return nil, errors.New("Parse Error")
		}

		seeds[name] = seed
	}

	return seeds, nil
}

type seedCollectionType int
const (
	seedInput seedCollectionType = iota
	seedOutput
	seedTable
	seedScratch
)

type table struct {
	key     []string
	columns []string
	source  source
	typ seedCollectionType
}

func newTable(typ seedCollectionType) *table {
	return &table{}
}

func (t *table) String() string {
	return fmt.Sprint(t.key, "=>", t.columns)
}

type tableCollection map[string]*table

func newTableCollection() tableCollection {
	return make(map[string]*table)
}
