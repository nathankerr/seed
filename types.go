package seed

// Seed is the datamodel for a Seed.
type Seed struct {
	Name        string
	Collections map[string]*Collection // string is same as collection.name
	Rules       []*Rule
}

// Collection describes the data managed in a service.
type Collection struct {
	Type CollectionType
	Key  []string
	Data []string
}

// CollectionType describes the type of the collection.
type CollectionType int

const (
	// CollectionInput collections receive data from outside the Seed
	CollectionInput CollectionType = iota

	// CollectionOutput collections transfer data from inside the Seed
	// to outside
	CollectionOutput

	// CollectionTable collections are persistent
	CollectionTable

	// CollectionScratch collections are not persistent and intended to
	// handle intermediate results within a single timestep
	CollectionScratch

	// CollectionChannel collections transfer data over the network.
	CollectionChannel
)

// A Rule describes the data manipulation possible in a service.
type Rule struct {
	Supplies  string
	Operation string
	Intension []Expression
	Predicate []Constraint
}

// Expression is the holder for all things that can be in a tuple.
// Use something like:
// switch value := expression.Value.(type) {
// case QualifiedColumn:
// case FunctionCall:
// default:
// 	panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression.Value).String()))
// }
type Expression interface{}

// MapFunction gathers the data to both represent a map function
// as well as a link to the actual function for execution.
type MapFunction struct {
	Name      string // the function name as a string, instead of as a function
	Function  MapFn
	Arguments []QualifiedColumn
}

// ReduceFunction gathers the data to both represent a reduce function
// as well as a link to the actual function for execution.
type ReduceFunction struct {
	Name      string
	Function  ReduceFn
	Arguments []QualifiedColumn
}

// Tuple is the set of values in a collection
type Tuple []interface{}

// Element is a element in tuple
type Element interface{}

// MapFn is the function type for a Map function
type MapFn func(Tuple) Element

// ReduceFn is the function type for a Reduce function
type ReduceFn func([]Tuple) Element

// QualifiedColumn uniquely identifies a column in a collection
// by name
type QualifiedColumn struct {
	Collection string
	Column     string
}

// Constraint identifies the columns which are to be held equivalent.
type Constraint struct {
	Left  QualifiedColumn
	Right QualifiedColumn
}
