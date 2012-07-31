package main

import (
	"fmt"
	// "log"
)

func parseinfo(args ...interface{}) {
	//log.Println(args...)
}

type seed struct {
	inputs  map[string]*table
	outputs map[string]*table
	tables  map[string]*table
	rules   []*rule
}

func (s *seed) String() string {
	str := "inputs:"
	for k, v := range s.inputs {
		str = fmt.Sprint(str, "\n\t", k, " ", v, "\t(", v.source, ")")
	}

	str += "\noutputs:"
	for k, v := range s.outputs {
		str = fmt.Sprint(str, "\n\t", k, " ", v, "\t(", v.source, ")")
	}

	str += "\ntables:"
	for k, v := range s.tables {
		str = fmt.Sprint(str, "\n\t", k, " ", v, "\t(", v.source, ")")
	}

	str += "\nrules:"
	for k, v := range s.rules {
		str = fmt.Sprint(str, "\n\t", k, " ", v, "\t(", v.source, ")")
	}

	return str
}

type table struct {
	key     []string
	columns []string
	source  source
}

func (t *table) String() string {
	return fmt.Sprint(t.key, "=>", t.columns)
}

type rule struct {
	value    string
	supplies []string
	requires []string
	source   source
}

func (r *rule) String() string {
	return r.value
}

func newSeed() *seed {
	return &seed{
		inputs:  make(map[string]*table),
		outputs: make(map[string]*table),
		tables:  make(map[string]*table),
	}
}

type parsefn func(p *parser) parsefn

type parser struct {
	s     *seed
	items chan item
	i     item // the last item
}

func (p *parser) nextItem() item {
	p.i = <-p.items
	return p.i
}

func parse(name, input string) *seed {
	p := &parser{}
	p.s = newSeed()

	l := lex(name, input)
	go l.run()

	p.items = l.items

	for state := parseSeed; state != nil; {
		state = state(p)
	}
	// log.Println("parser stopped running")

	return p.s
}

func parseSeed(p *parser) parsefn {
	parseinfo("parseSeed")

	i := p.nextItem()

	switch i.typ {
	case itemInput:
		return parseInput
	case itemOutput:
		return parseOutput
	case itemTable:
		return parseTable
	case itemIdentifier:
		return parseRule
	case itemEOF:
		return nil
	default:
		fmt.Println("parseSeed: unexpected", i)
		return nil
	}

	return nil
}

// input <name> <schema>
func parseInput(p *parser) parsefn {
	parseinfo("parseInput")

	i := p.nextItem()
	if i.typ != itemIdentifier {
		fmt.Println("parseInput: expected itemIdentifier, got ", i)
		return nil
	}

	name := i.val

	schema, ok := parseSchema(p.items)
	if !ok {
		return nil
	}

	if _, ok := p.s.inputs[name]; ok {
		fmt.Println("parseInput: input", name, "already exists")
		return nil
	}

	schema.source = i.source
	p.s.inputs[name] = schema

	return parseSeed
}

// [key] [columns]
func parseSchema(items chan item) (schema *table, ok bool) {
	parseinfo("parseSchema")

	schema = new(table)

	key, ok := parseArray(items)
	if !ok {
		fmt.Println("parseSchema: expected key array")
		return nil, false
	}
	schema.key = key

	columns, ok := parseArray(items)
	if !ok {
		fmt.Println("parseSchema: expected columns array")
		return nil, false
	}
	schema.columns = columns

	return schema, true
}

// [string, string]
func parseArray(items chan item) (array []string, ok bool) {
	parseinfo("parseArray")

	i := <-items
	if i.typ != itemBeginArray {
		fmt.Println("parseSchema: expected [, got", i.val)
		return nil, false
	}

	i = <-items
	if i.typ != itemIdentifier {
		fmt.Println("parseSchema: expected identifier, got", i.val)
		return nil, false
	}
	array = append(array, i.val)

	i = <-items
	for {
		switch i.typ {
		case itemEndArray:
			return array, true
		case itemArrayDelimter:
			i = <-items
			if i.typ != itemIdentifier {
				fmt.Println("parseSchema: expected identifier, got", i.val)
				return nil, false
			}
			array = append(array, i.val)
		}
		i = <-items
	}
	return nil, false
}

// output <name> <schema>
func parseOutput(p *parser) parsefn {
	parseinfo("parseOutput")

	i := p.nextItem()
	if i.typ != itemIdentifier {
		fmt.Println("parseOutput: expected itemIdentifier, got ", i)
		return nil
	}

	name := i.val

	schema, ok := parseSchema(p.items)
	if !ok {
		return nil
	}

	if _, ok := p.s.inputs[name]; ok {
		fmt.Println("parseOutput: input", name, "already exists")
		return nil
	}

	schema.source = i.source
	p.s.outputs[name] = schema

	return parseSeed
}

// table <name> <schema>
func parseTable(p *parser) parsefn {
	parseinfo("parseTable")

	i := p.nextItem()
	if i.typ != itemIdentifier {
		fmt.Println("parseTable: expected itemIdentifier, got ", i)
		return nil
	}

	name := i.val

	schema, ok := parseSchema(p.items)
	if !ok {
		return nil
	}

	if _, ok := p.s.inputs[name]; ok {
		fmt.Println("parseTable: input", name, "already exists")
		return nil
	}

	schema.source = i.source
	p.s.tables[name] = schema

	return parseSeed
}

func parseRule(p *parser) parsefn {
	parseinfo("parseRule")

	// destination
	destination := p.i

	// operation
	operation := p.nextItem()
	if operation.typ != itemOperationInsert {
		fmt.Println("parseRule: expected itemOperationInsert, got ", operation)
		return nil
	}

	// expression
	expr := p.nextItem()
	if expr.typ != itemIdentifier {
		fmt.Println("parseRule: expected itemIdentifier, got ", expr)
		return nil
	}

	r := new(rule)
	r.value = fmt.Sprint(destination.val, " ", operation.val, " ", expr.val)
	r.supplies = append(r.supplies, destination.val)
	r.requires = append(r.requires, expr.val)
	r.source = destination.source
	p.s.rules = append(p.s.rules, r)

	return parseSeed
}
