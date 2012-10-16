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
	// lhs
	supplies string

	// op
	typ      ruleType

	//rhs
	collections map[string]bool // boolean has no meaning, just want a map for unique keys
	output []qualifiedColumn
	predicates []predicate

	// meta
	requires []string
	source   source
}

func newRule(src source) *rule {
	collections := make(map[string]bool)
	return &rule{source: src, collections: collections}
}

func (r *rule) String() string {
	output := []string{}
	for _, o := range r.output {
		output = append(output, o.String())
	}

	if len(r.predicates) > 0 {
		predicates := []string{}
		for _, p := range r.predicates {
			predicates = append(predicates, p.String())
		}
		return fmt.Sprintf("[%s]: %s",
			strings.Join(output, ", "),
			strings.Join(predicates, ", "))
	}

	return fmt.Sprintf("%s %s [%s]",
		r.supplies,
		r.typ.String(),
		strings.Join(output, ", "))
}

type Rubyer interface {
	Ruby() string
}

func (r *rule) Ruby() string {
	var str string

	collections := []string{}
	for c, _ := range r.collections {
		collections = append(collections, c)
	}

	index := make(map[string]string)
	names := []string{}
	for i, c := range collections {
		name := fmt.Sprintf("c%d", i)
		index[c] = name
		names = append(names, name)
	}

	output := []string{}
	for _, o := range r.output {
		output = append(output, fmt.Sprintf("%s.%s", index[o.collection], o.column))
	}

	if len(r.collections) == 1 {
		str = fmt.Sprintf("%s do |%s| [%s] end",
			r.output[0].collection,
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	} else {
		predicates := []string{}
		for _, p := range r.predicates {
			predicates = append(predicates, p.String())
		}
		
		str = fmt.Sprintf("(%s).combos(%s) do |%s| [%s] end",
			strings.Join(collections, " * "),
			strings.Join(predicates, ", "),
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	}

	return fmt.Sprintf("%s %s %s",
		r.supplies,
		r.typ,
		str)
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

	if len(j.collections) == 1 {
		return fmt.Sprintf("%s do |%s| [%s] end",
			j.output[0].collection,
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	} else {
		predicates := []string{}
		for _, p := range j.predicates {
			predicates = append(predicates, p.String())
		}
		
		return fmt.Sprintf("(%s).combos(%s) do |%s| [%s] end",
			strings.Join(collections, " * "),
			strings.Join(predicates, ", "),
			strings.Join(names, ", "),
			strings.Join(output, ", "))
	}

	panic("shouldn't get here")
	return ""
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