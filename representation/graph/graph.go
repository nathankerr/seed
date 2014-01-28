// Represents Seeds as directed graphs
package graph

import (
	"fmt"
	"github.com/nathankerr/graph"
	"github.com/nathankerr/seed"
	"reflect"
)

// used to implement graph.Graph
type Graph struct {
	*seed.Seed
	nodes   []graph.Node   // used to map graph.Node (which uses Int)
	nodeFor map[string]int // maps collection names and rule numbers to index in nodes
}

func SeedAsGraph(seed *seed.Seed) *Graph {
	g := &Graph{
		Seed:    seed,
		nodes:   make([]graph.Node, len(seed.Collections)+len(seed.Rules)),
		nodeFor: map[string]int{},
	}

	id := 0
	for collectionName, collection := range seed.Collections {
		g.nodes[id] = CollectionNode{
			id:         id,
			Name:       collectionName,
			Collection: collection,
		}
		g.nodeFor[collectionName] = id
		id++
	}

	for ruleNumber, rule := range seed.Rules {
		g.nodes[id] = RuleNode{
			id:   id,
			num:  ruleNumber,
			Rule: rule,
		}
		g.nodeFor[fmt.Sprint(ruleNumber)] = id
		id++
	}

	return g
}

func (g *Graph) GetNode(id int) graph.Node {
	return g.nodes[id]
}

// name is collection name or rule number (as string)
func (g *Graph) NodeFor(name string) (graph.Node, bool) {
	id, ok := g.nodeFor[name]
	if !ok {
		return nil, false
	}

	return g.nodes[id], true
}

// Gives the nodes connected by OUTBOUND edges
// used by A*
func (g Graph) Successors(node graph.Node) []graph.Node {
	successors := []graph.Node{}

	switch node := node.(type) {
	case CollectionNode:
		for ruleNumber, rule := range g.Seed.Rules {
			isSuccessor := false

			for _, collectionName := range rule.Requires() {
				if collectionName == node.Name {
					isSuccessor = true
					break
				}
			}

			if isSuccessor {
				successor, ok := g.NodeFor(fmt.Sprint(ruleNumber))
				if ok {
					successors = append(successors, successor)
				}
			}
		}
	case RuleNode:
		successor, ok := g.NodeFor(node.Rule.Supplies)
		if ok {
			successors = append(successors, successor)
		}
	default:
		panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(node).String()))
	}

	return successors
}

func (g Graph) IsSuccessor(node, successor graph.Node) bool {
	panic("TODO")
}

func (g Graph) Predecessors(node graph.Node) []graph.Node {
	panic("TODO")
}

func (g Graph) IsPredecessor(node, predecessor graph.Node) bool {
	panic("TODO")
}

func (g Graph) IsAdjacent(node, neighbor graph.Node) bool {
	panic("TODO")
}

func (g Graph) NodeExists(node graph.Node) bool {
	panic("TODO")
}

func (g Graph) Degree(node graph.Node) int {
	panic("TODO")
}

func (g Graph) EdgeList() []graph.Edge {
	edges := []graph.Edge{}

	for _, node := range g.NodeList() {
		for _, successor := range g.Successors(node) {
			edges = append(edges, Edge{
				From: node,
				To:   successor,
			})
		}
	}

	return edges
}

// for GraphToDot
func (g Graph) NodeList() []graph.Node {
	return g.nodes
}

func (g Graph) IsDirected() bool {
	return true
}

type CollectionNode struct {
	id   int
	Name string
	*seed.Collection
}

func (cnode CollectionNode) ID() int {
	return cnode.id
}

type RuleNode struct {
	id  int
	num int
	*seed.Rule
}

func (rnode RuleNode) ID() int {
	return rnode.id
}

type Edge struct {
	From, To graph.Node
}

func (edge Edge) Head() graph.Node {
	return edge.From
}

func (edge Edge) Tail() graph.Node {
	return edge.To
}
