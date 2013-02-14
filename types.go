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
	Projection []Expression
	Predicate  []Constraint
	Source     Source
}

// Use something like:
// switch value := expression.Value.(type) {
// case QualifiedColumn:
// case FunctionCall:
// default:
// 	panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression.Value).String()))
// }
type Expression struct {
	Value interface{} // QualifiedColumn, FunctionCall
}

type FunctionCall struct {
	Name      string      // the function name as a string, instead of as a function
	Function  interface{} // MapFn, ReduceFn
	Arguments []QualifiedColumn
}

type Tuple []interface{}
type MapFn func(Tuple) Tuple
type ReduceFn func([]Tuple) Tuple

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
