package main

type service struct {
	collections map[string]*collection // string is same as collection.name
	rules       []*rule
	source      source
}

type collection struct {
	ctype  collectionType
	key    []string
	data   []string
	source source
}

type collectionType int

const (
	collectionInput collectionType = iota
	collectionOutput
	collectionTable
	collectionScratch // only for use in buds
	collectionChannel // only for use in buds
)

type rule struct {
	supplies   string
	operation  string
	projection []qualifiedColumn
	predicate  []constraint
	block      string
	source     source
}

type qualifiedColumn struct {
	collection string
	column     string
	source     source
}

type constraint struct {
	left   qualifiedColumn
	right  qualifiedColumn
	source source
}

type source struct {
	name   string
	line   int
	column int
}

type group struct {
	rules       []int
	collections map[string]collectionType
}
