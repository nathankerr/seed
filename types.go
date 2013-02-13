package seed

type Seed struct {
	Collections map[string]*Collection // string is same as collection.name
	Rules       []*Rule
	Source      Source
}

type Collection struct {
	Type   CollectionType
	Key    []string
	Data   []string
	Source Source
}

type CollectionType int

const (
	CollectionInput CollectionType = iota
	CollectionOutput
	CollectionTable
	CollectionScratch // only for use in buds
	CollectionChannel // only for use in buds
)

type Rule struct {
	Supplies   string
	Operation  string
	Projection []QualifiedColumn
	Predicate  []Constraint
	Source     Source
}

type QualifiedColumn struct {
	Collection string
	Column     string
}

type Constraint struct {
	Left  QualifiedColumn
	Right QualifiedColumn
}

type Source struct {
	Name   string
	Line   int
	Column int
}
