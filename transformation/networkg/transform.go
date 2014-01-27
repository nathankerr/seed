// Package networkg uses graphs to add network interfaces to full (non-subset) Seed programs
package networkg

import (
	"fmt"
	"github.com/nathankerr/graph"
	"github.com/nathankerr/seed"
	seedGraph "github.com/nathankerr/seed/representation/graph"
	"math"
	"reflect"
	"strings"
)

// Transform uses graphs to add network interfaces to full (non-subset) Seeds
func Transform(orig *seed.Seed) (*seed.Seed, error) {
	// build graph
	g := seedGraph.SeedAsGraph(orig)

	orig.Name = strings.Title(orig.Name) + "Server"

	for inputName, input := range orig.Collections {
		if input.Type != seed.CollectionInput {
			continue
		}
		input.Key = append([]string{"@address"}, input.Key...)

		start, ok := g.NodeFor(inputName)
		if !ok {
			panic("could not find node for " + inputName)
		}

		for outputName, output := range orig.Collections {
			if output.Type != seed.CollectionOutput {
				continue
			}
			goal, ok := g.NodeFor(outputName)
			if !ok {
				panic("could not find node for " + outputName)
			}
			path, cost, _ := graph.AStar(start, goal, g, cost, nil)
			if math.IsInf(cost, 0) {
				// there is no path (that does not go through a table)
				continue
			}

			outputAddress := outputName + "_addr"
			previousCollection := inputName
			for _, node := range path[:len(path)] {
				// unbox graph.internalNode;
				node = g.GetNode(node.ID())

				switch node := node.(type) {
				case seedGraph.CollectionNode:
					previousCollection = node.Name
					switch node.Collection.Type {
					case seed.CollectionInput, seed.CollectionScratch:
						node.Collection.Key = prependIfNotExists(node.Collection.Key, outputAddress)
					case seed.CollectionOutput:
						node.Collection.Key = prependIfNotExists(node.Collection.Key, "@"+outputAddress)
					case seed.CollectionChannel, seed.CollectionTable:
						panic("should not encounter these collection types")
					default:
						panic(fmt.Sprintf("unhandled type: %d", node.Collection.Type))
					}
				case seedGraph.RuleNode:
					exists := false
					for _, expression := range node.Rule.Projection {
						switch expression := expression.(type) {
						case seed.QualifiedColumn:
							if expression.Column == outputAddress {
								// if a reference to the outputAddress already exists (i.e., from being added by another flow)
								// then add a constraint to make the rows match up
								node.Rule.Predicate = append(node.Rule.Predicate, seed.Constraint{
									Left: expression,
									Right: seed.QualifiedColumn{
										Collection: previousCollection,
										Column:     outputAddress,
									},
								})
								exists = true
							}
						case seed.MapFunction, seed.ReduceFunction:
							continue
						default:
							panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression).String()))
						}
					}
					if !exists {
						// add to the projection
						node.Rule.Projection = append([]seed.Expression{seed.QualifiedColumn{
							Collection: previousCollection,
							Column:     outputAddress,
						}}, node.Rule.Projection...)
					}
				default:
					panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(node).String()))
				}
			}
		}
	}

	// change inputs and outputs to channels
	for _, collection := range orig.Collections {
		switch collection.Type {
		case seed.CollectionInput, seed.CollectionOutput:
			collection.Type = seed.CollectionChannel
		case seed.CollectionScratch, seed.CollectionTable, seed.CollectionChannel:
			// no-op
		default:
			panic(collection.Type)
		}
	}

	// rules supplying channels must be asynchronous
	for _, rule := range orig.Rules {
		collectionName := rule.Supplies
		collection, ok := orig.Collections[collectionName]
		if !ok {
			// should never happen
			panic(collectionName)
		}

		if collection.Type == seed.CollectionChannel {
			rule.Operation = "<~"
		}
	}

	return orig, nil
}

// returns Inf if the to node is a table, otherwise 1
func cost(from graph.Node, to graph.Node) float64 {
	switch to := to.(type) {
	case seedGraph.CollectionNode:
		if to.Collection.Type == seed.CollectionTable {
			return math.Inf(0)
		}
	case seedGraph.RuleNode:
		// no-op
	default:
		panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(to).String()))
	}

	return 1.0
}

func prependIfNotExists(strings []string, toAdd string) []string {
	exists := false

	for _, str := range strings {
		if str == toAdd {
			exists = true
		}
	}

	if !exists {
		strings = append([]string{toAdd}, strings...)
	}

	return strings
}
