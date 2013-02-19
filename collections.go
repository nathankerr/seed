package seed

import (
	"fmt"
	"reflect"
)

func (r *Rule) Collections() []string {
	collectionsmap := make(map[string]bool) // map only used for uniqueness

	// supplies
	collectionsmap[r.Supplies] = true

	for _, requires := range r.Requires() {
		collectionsmap[requires] = true
	}

	// convert map to []string
	collections := []string{}
	for collection, _ := range collectionsmap {
		collections = append(collections, collection)
	}

	return collections
}

func (r *Rule) Requires() []string {
	requiresmap := make(map[string]bool) // map only used for uniqueness

	// projection
	for _, expression := range r.Projection {
		switch value := expression.Value.(type) {
		case QualifiedColumn:
			requiresmap[value.Collection] = true
		case MapFunction:
			for _, qc := range value.Arguments {
				requiresmap[qc.Collection] = true
			}
		default:
			panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression.Value).String()))
		}
	}

	// predicate
	for _, c := range r.Predicate {
		requiresmap[c.Left.Collection] = true
		requiresmap[c.Right.Collection] = true
	}

	// convert map to []string
	requires := []string{}
	for collection, _ := range requiresmap {
		requires = append(requires, collection)
	}

	return requires
}
