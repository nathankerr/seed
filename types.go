package main

type service struct {
	Collections map[string]*collection // string is same as collection.name
	Rules       []*rule
	Source      source
}

type collection struct {
	Type   collectionType
	Key    []string
	Data   []string
	Source source
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
	Supplies   string
	Operation  string
	Projection []qualifiedColumn
	Predicate  []constraint
	Block      string
	Source     source
}

type qualifiedColumn struct {
	Collection string
	Column     string
}

type constraint struct {
	Left  qualifiedColumn
	Right qualifiedColumn
}

type source struct {
	Name   string
	Line   int
	Column int
}
