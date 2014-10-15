package seed

import (
	"fmt"
	"reflect"
)

// AddressColumn determines which column is used for addresses.
func (c *Collection) AddressColumn() (int, bool) {
	addressColumn := -1
	ok := false

	if c.Type != CollectionChannel {
		return -1, false
	}

	for index, name := range c.Key {
		if name[0] == '@' {
			addressColumn = index
			break
		}
	}
	if addressColumn >= 0 {
		ok = true
	}

	return addressColumn, ok
}

// Collections returns the list of collections interacted with by the rule.
func (r *Rule) Collections() []string {
	collectionsmap := make(map[string]bool) // map only used for uniqueness

	// supplies
	collectionsmap[r.Supplies] = true

	for _, requires := range r.Requires() {
		collectionsmap[requires] = true
	}

	// convert map to []string
	collections := []string{}
	for collection := range collectionsmap {
		collections = append(collections, collection)
	}

	return collections
}

// Requires returns the set of collections that the rule gets data from.
func (r *Rule) Requires() []string {
	requiresmap := make(map[string]bool) // map only used for uniqueness

	// intension
	for _, expression := range r.Intension {
		switch value := expression.(type) {
		case QualifiedColumn:
			requiresmap[value.Collection] = true
		case MapFunction:
			for _, qc := range value.Arguments {
				requiresmap[qc.Collection] = true
			}
		case ReduceFunction:
			for _, qc := range value.Arguments {
				requiresmap[qc.Collection] = true
			}
		default:
			panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression).String()))
		}
	}

	// predicate
	for _, c := range r.Predicate {
		requiresmap[c.Left.Collection] = true
		requiresmap[c.Right.Collection] = true
	}

	// convert map to []string
	requires := []string{}
	for collection := range requiresmap {
		requires = append(requires, collection)
	}

	return requires
}
