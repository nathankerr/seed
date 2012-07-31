package main

import (
	"fmt"
	"log"
)

func parseinfo(args ...interface{}) {
	//log.Println(args...)
}

type seed struct {
	inputs  map[string]*tableSchema
	outputs map[string]*tableSchema
	tables  map[string]*tableSchema
	rules   []string
}

func (s *seed) String() string {
	return fmt.Sprint("inputs: ", s.inputs, "\noutputs: ", s.outputs, "\ntables: ", s.tables, "\nrules: ", s.rules)
}

func (s* seed) addRule(destination, operation, source string) {
	rule := fmt.Sprint(destination, " ", operation, " ", source)
	s.rules = append(s.rules, rule)
}

type tableSchema struct {
	key     []string
	columns []string
}

func (t* tableSchema) String() string {
	return fmt.Sprint(t.key, "=>", t.columns)
}

func newSeed() *seed {
	return &seed{
		inputs:  make(map[string]*tableSchema),
		outputs: make(map[string]*tableSchema),
		tables:  make(map[string]*tableSchema),
	}
}

type parsefn func(p *parser) parsefn

type parser struct {
	s *seed
	items chan item
	i item // the last item
}

func (p *parser) nextItem() item {
	p.i = <- p.items
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
	log.Println("parser stopped running")

	fmt.Println(p.s)

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
	p.s.inputs[name] = schema

	return parseSeed
}

// [key] [columns]
func parseSchema(items chan item) (schema *tableSchema, ok bool) {
	parseinfo("parseSchema")

	schema = new(tableSchema)

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

	i := <- items
	if i.typ != itemBeginArray {
		fmt.Println("parseSchema: expected [, got", i.val)
		return nil, false
	}

	i = <- items
	if i.typ != itemIdentifier {
		fmt.Println("parseSchema: expected identifier, got", i.val)
		return nil, false
	}
	array = append(array, i.val)

	i = <- items
	for {
		switch i.typ {
		case itemEndArray:
			return array, true
		case itemArrayDelimter:
			i = <- items
			if i.typ != itemIdentifier {
				fmt.Println("parseSchema: expected identifier, got", i.val)
				return nil, false
			}
			array = append(array, i.val)
		}
		i = <- items
	}
	return nil, false
}

// output <name> <schema>
func parseOutput(p * parser) parsefn {
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

	// source
	source := p.nextItem()
	if source.typ != itemIdentifier {
		fmt.Println("parseRule: expected itemIdentifier, got ", source)
		return nil
	}

	p.s.addRule(destination.val, operation.val, source.val)

	return parseSeed
}