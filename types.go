package seed

type Seed struct {
	Name        string
	Collections map[string]*Collection // string is same as collection.name
	Rules       []*Rule
}

type Collection struct {
	Type CollectionType
	Key  []string
	Data []string
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
	Supplies  string
	Operation string
	Intension []Expression
	Predicate []Constraint
}

// Use something like:
// switch value := expression.Value.(type) {
// case QualifiedColumn:
// case FunctionCall:
// default:
// 	panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression.Value).String()))
// }
type Expression interface{}

type MapFunction struct {
	Name      string // the function name as a string, instead of as a function
	Function  MapFn
	Arguments []QualifiedColumn
}

type ReduceFunction struct {
	Name      string
	Function  ReduceFn
	Arguments []QualifiedColumn
}

type Tuple []interface{}
type Element interface{}
type MapFn func(Tuple) Element
type ReduceFn func([]Tuple) Element

type QualifiedColumn struct {
	Collection string
	Column     string
}

type Constraint struct {
	Left  QualifiedColumn
	Right QualifiedColumn
}
