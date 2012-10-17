package main

import (
	"fmt"
	"strings"
)

type ruleType int

const (
	ruleInsert ruleType = iota
	ruleDelete
	ruleUpdate
	ruleAsyncInsert // only to be used with bud
)

var ruleTypeNames = map[ruleType]string{
	ruleInsert:      "<+",
	ruleDelete:      "<-",
	ruleUpdate:      "<+-",
	ruleAsyncInsert: "<~",
}

func (rt ruleType) String() string {
	str, ok := ruleTypeNames[rt]
	if !ok {
		panic("unknown rule type")
	}
	return str
}

type rule struct {
	// lhs
	supplies string

	// op
	typ ruleType

	//rhs
	output     []qualifiedColumn
	predicates []predicate

	// meta
	requires map[string]bool // bool has no meaning, just want a map for unique keys
	source   source
}

func newRule(src source) *rule {
	requires := make(map[string]bool)
	return &rule{source: src, requires: requires}
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
	for c, _ := range r.requires {
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

	if len(r.requires) == 1 {
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

type qualifiedColumn struct {
	collection string
	column     string
}

func (qc *qualifiedColumn) String() string {
	return fmt.Sprintf("%s.%s", qc.collection, qc.column)
}

type predicate struct {
	left  qualifiedColumn
	right qualifiedColumn
}

func (p *predicate) String() string {
	return fmt.Sprintf("%s => %s", p.left.String(), p.right.String())
}
