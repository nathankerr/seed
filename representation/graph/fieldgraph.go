// Represents Seeds as directed graphs
package graph

import (
	"fmt"
	"github.com/nathankerr/graph"
	"github.com/nathankerr/seed"
	"reflect"
	// "strings"
)

// used to implement graph.FieldGraph
type FieldGraph struct {
	*seed.Seed
	nodes   []graph.Node   // used to map graph.Node (which uses Int)
	nodeFor map[string]int // maps qualified column names and rule numbers to index in nodes
}

func SeedAsFieldGraph(seed *seed.Seed) *FieldGraph {
	g := &FieldGraph{
		Seed:    seed,
		nodes:   []graph.Node{},
		nodeFor: map[string]int{},
	}

	id := 0
	for collectionName, collection := range seed.Collections {
		for _, column := range append(collection.Key, collection.Data...) {
			columnName := fmt.Sprintf("%s.%s", collectionName, column)
			g.nodes = append(g.nodes, FieldNode{
				id:             id,
				collectionName: collectionName,
				field:          column,
			})
			g.nodeFor[columnName] = id
			id++
		}
	}

	return g
}

func (g *FieldGraph) GetNode(id int) graph.Node {
	return g.nodes[id]
}

// name is collection name or rule number (as string)
func (g *FieldGraph) NodeFor(name string) (graph.Node, bool) {
	id, ok := g.nodeFor[name]
	if !ok {
		return nil, false
	}

	return g.nodes[id], true
}

// Gives the nodes connected by OUTBOUND edges
func (g FieldGraph) Successors(node graph.Node) []graph.Node {
	panic("TODO")
}

func (g FieldGraph) IsSuccessor(node, successor graph.Node) bool {
	panic("TODO")
}

func (g FieldGraph) Predecessors(node graph.Node) []graph.Node {
	panic("TODO")
}

func (g FieldGraph) IsPredecessor(node, predecessor graph.Node) bool {
	panic("TODO")
}

func (g FieldGraph) IsAdjacent(node, neighbor graph.Node) bool {
	panic("TODO")
}

func (g FieldGraph) NodeExists(node graph.Node) bool {
	panic("TODO")
}

func (g FieldGraph) Degree(node graph.Node) int {
	panic("TODO")
}

func (g FieldGraph) EdgeList() []graph.Edge {
	edges := []graph.Edge{}

	for ruleNumber, rule := range g.Seed.Rules {
		arcs := map[seed.QualifiedColumn]seed.QualifiedColumn{}

		equivalentFields := map[seed.QualifiedColumn]seed.QualifiedColumn{}
		for _, constraint := range rule.Predicate {
			equivalentFields[constraint.Left] = constraint.Right
			equivalentFields[constraint.Right] = constraint.Left
		}

		for columnNumber, expression := range rule.Projection {
			switch expression := expression.(type) {
			case seed.QualifiedColumn:
				arcs[expression] = g.qcFor(ruleNumber, columnNumber)

				if equivalentField, ok := equivalentFields[expression]; ok {
					arcs[equivalentField] = g.qcFor(ruleNumber, columnNumber)
				}
			case seed.MapFunction:
				for _, expression := range expression.Arguments {
					arcs[expression] = g.qcFor(ruleNumber, columnNumber)

					if equivalentField, ok := equivalentFields[expression]; ok {
						arcs[equivalentField] = g.qcFor(ruleNumber, columnNumber)
					}
				}
			case seed.ReduceFunction:
				for _, expression := range expression.Arguments {
					arcs[expression] = g.qcFor(ruleNumber, columnNumber)

					if equivalentField, ok := equivalentFields[expression]; ok {
						arcs[equivalentField] = g.qcFor(ruleNumber, columnNumber)
					}
				}
			default:
				panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression).String()))
			}
		}

		for from, to := range arcs {
			nodeFrom, ok := g.NodeFor(from.String())
			if !ok {
				panic(from.String())
			}

			nodeTo, ok := g.NodeFor(to.String())
			if !ok {
				panic(from.String())
			}

			edges = append(edges, FieldEdge{
				Edge: Edge{
					From: nodeFrom,
					To:   nodeTo,
				},
				Label: fmt.Sprintf("rule %d", ruleNumber),
			})
		}
	}

	return edges
}

// for dot
func (g FieldGraph) NodeList() []graph.Node {
	return g.nodes
}

func (g FieldGraph) IsDirected() bool {
	return true
}

type FieldNode struct {
	id             int
	collectionName string
	field          string
}

func (cnode FieldNode) ID() int {
	return cnode.id
}

type FieldEdge struct {
	Edge
	Label string
}

func (g *FieldGraph) qcFor(ruleNumber, columnNumber int) seed.QualifiedColumn {
	suppliesName := g.Seed.Rules[ruleNumber].Supplies
	supplies := g.Seed.Collections[suppliesName]

	columnNames := append(supplies.Key, supplies.Data...)

	return seed.QualifiedColumn{
		Collection: g.Seed.Rules[ruleNumber].Supplies,
		Column:     columnNames[columnNumber],
	}
}
