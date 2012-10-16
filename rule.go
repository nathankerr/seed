package main

import(
	"fmt"
	"strings"
)

type ruleType int

const (
	ruleInsert ruleType = iota
	ruleSet
	ruleDelete
	ruleUpdate
	ruleAsyncInsert // only to be used with bud
)

func (rt ruleType) String() string {
	switch rt {
	case ruleInsert:
		return "<+"
	case ruleSet:
		return "<="
	case ruleDelete:
		return "<-"
	case ruleUpdate:
		return "<+-"
	case ruleAsyncInsert:
		return "<~"
	default:
		panic("unknown type")
	}
	return "ERROR"
}

type rule struct {
	value    interface{}
	typ      ruleType
	supplies string
	requires []string
	source   source
}

func newRule(src source) *rule {
	return &rule{source: src}
}

func (r *rule) String() string {
	switch value := r.value.(type) {
	case fmt.Stringer:
		return fmt.Sprintf("%s %s %s",
			r.supplies,
			r.typ.String(),
			value)
	case string:
		return value
	default:
		panic("Unsupported type")
	}
	return "Unsupported type: see rule.go r.String()"
}

type Rubyer interface {
	Ruby() string
}

func (r *rule) Ruby() string {
	switch value := r.value.(type) {
	case Rubyer:
		return fmt.Sprintf("%s %s %s",
			r.supplies,
			r.typ,
			value.Ruby())
	case string:
		return value
	default:
		panic("Unsupported type")
	}
	return "Unsupported type: see rule.go r.String()"
}

type join struct {
	collections map[string]bool // boolean has no meaning, just want a map for unique keys
	output []qualifiedColumn
	predicates []predicate
}

func newJoin() *join {
	collections := make(map[string]bool)
	return &join{collections: collections}
}

func (j *join) String() string {
	output := []string{}
	for _, o := range j.output {
		output = append(output, o.String())
	}

	if len(j.predicates) > 0 {
		predicates := []string{}
		for _, p := range j.predicates {
			predicates = append(predicates, p.String())
		}
		return fmt.Sprintf("[%s]: %s",
			strings.Join(output, ", "),
			strings.Join(predicates, ", "))
	}

	return fmt.Sprintf("[%s]", strings.Join(output, ", "))
}

func (j *join) Ruby() string {
	collections := []string{}
	for c, _ := range j.collections {
		collections = append(collections, c)
	}

	predicates := []string{}
	for _, p := range j.predicates {
		predicates = append(predicates, p.String())
	}

	index := make(map[string]string)
	names := []string{}
	for i, c := range collections {
		name := fmt.Sprintf("c%d", i)
		index[c] = name
		names = append(names, name)
	}

	output := []string{}
	for _, o := range j.output {
		output = append(output, fmt.Sprintf("%s.%s", index[o.collection], o.column))
	}
	
	return fmt.Sprintf("(%s).combos(%s) do |%s| [%s] end",
		strings.Join(collections, " * "),
		strings.Join(predicates, ", "),
		strings.Join(names, ", "),
		strings.Join(output, ", "))
}

type qualifiedColumn struct {
	collection string
	column string
}

func(qc *qualifiedColumn) String() string {
	return fmt.Sprintf("%s.%s", qc.collection, qc.column)
}

type predicate struct {
	left qualifiedColumn
	right qualifiedColumn
}

func (p *predicate) String() string {
	return fmt.Sprintf("%s => %s", p.left.String(), p.right.String())
}