package graph

import (
	"fmt"
	"github.com/nathankerr/graph"
	"github.com/nathankerr/seed"
)

func ToGraph(seed *seed.Seed, name string) ([]byte, error) {
	return GraphToDot(SeedAsGraph(seed), name)
}

func ToFieldGraph(seed *seed.Seed, name string) ([]byte, error) {
	return GraphToDot(SeedAsFieldGraph(seed), name)
}

func GraphToDot(graph graph.Graph, name string) ([]byte, error) {
	dot := fmt.Sprintf("digraph %s {", name)
	dot = fmt.Sprintf("%s\n\tmargin=\"0\"", dot)
	dot = fmt.Sprintf("%s\n", dot)

	for _, node := range graph.NodeList() {
		nodeName := nameFor(node)
		nodeLabel := labelFor(node)
		dot = fmt.Sprintf("%s\n\t%s [label=\"%s\"]", dot, nodeName, nodeLabel)
	}

	for _, edge := range graph.EdgeList() {
		from := nameFor(edge.Head())
		to := nameFor(edge.Tail())

		label := ""
		switch edge := edge.(type) {
		case FieldEdge:
			label = fmt.Sprintf("[label=\"%s\"]", edge.Label)
		}

		dot = fmt.Sprintf("%s\n\t%s -> %s %s", dot, from, to, label)
	}

	return []byte(fmt.Sprintf("%s\n}", dot)), nil
}

func labelFor(node graph.Node) string {
	switch node := node.(type) {
	case CollectionNode:
		return node.Name
	case RuleNode:
		return fmt.Sprintf("rule %d", node.num)
	case FieldNode:
		return fmt.Sprintf("%s.%s", node.collectionName, node.field)
	default:
		return fmt.Sprint(node.ID())
	}
}

func nameFor(node graph.Node) string {
	switch node := node.(type) {
	case CollectionNode:
		return node.Name
	case RuleNode:
		return fmt.Sprintf("rule%d", node.num)
	case FieldNode:
		fieldName := node.field
		if fieldName[0] == '@' {
			fieldName = node.field[1:]
		}
		return fmt.Sprintf("%s_%s", node.collectionName, fieldName)
	default:
		return fmt.Sprint(node.ID())
	}
}
