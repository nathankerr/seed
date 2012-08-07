package main

import (
	"fmt"
	// "log"
	"os"
	"path"
	"runtime"
)

func parseinfo(args ...interface{}) {
	// log.Println(args...)
}

func parseError(i item, args ...interface{}) {
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		name := path.Ext(runtime.FuncForPC(pc).Name())
		name = name[1:]
		file = path.Base(file)
		fmt.Fprintf(os.Stderr, "%s:%d: Parse Error:\n", file, line)
		fmt.Fprintf(os.Stderr, "%s: %s: ", i.source, name)
	} else {
		fmt.Fprintf(os.Stderr, "%s: ", i.source)
	}
	fmt.Fprintln(os.Stderr, args...)
}

type parsefn func(p *parser) (next parsefn, ok bool)

type parser struct {
	s     *seed
	items chan item
	i     item // the last item
}

func (p *parser) nextItem() item {
	p.i = <-p.items
	return p.i
}

func parse(name, input string) (s *seed, ok bool) {
	p := &parser{}
	p.s = newSeed()

	l := lex(name, input)
	go l.run()

	p.items = l.items

	for state := parseSeed; state != nil; {
		state, ok = state(p)
		if !ok {
			return nil, false
		}
	}
	// log.Println("parser stopped running")

	return p.s, true
}

func parseSeed(p *parser) (next parsefn, ok bool) {
	parseinfo("parseSeed")

	i := p.nextItem()

	switch i.typ {
	case itemInput:
		return parseInput, true
	case itemOutput:
		return parseOutput, true
	case itemTable:
		return parseTable, true
	case itemIdentifier:
		return parseRule, true
	case itemEOF:
		return nil, true
	default:
		parseError(i, "unexpected", i)
		return nil, false
	}

	return nil, true
}

// input <name> <schema>
func parseInput(p *parser) (next parsefn, ok bool) {
	parseinfo("parseInput")

	i := p.nextItem()
	if i.typ != itemIdentifier {
		parseError(i, "expected itemIdentifier, got ", i)
		return nil, false
	}

	name := i.val

	schema, ok := parseSchema(p.items)
	if !ok {
		return nil, false
	}

	if _, ok := p.s.inputs[name]; ok {
		parseError(i, "input", name, "already exists")
		return nil, false
	}

	schema.source = i.source
	p.s.inputs[name] = schema

	return parseSeed, true
}

// [key] [columns]
func parseSchema(items chan item) (schema *table, ok bool) {
	parseinfo("parseSchema")

	schema = new(table)

	key, i, ok := parseArray(items)
	if !ok {
		parseError(i, "expected key array")
		return nil, false
	}
	schema.key = key

	columns, i, ok := parseArray(items)
	if !ok {
		parseError(i, "expected columns array")
		return nil, false
	}
	schema.columns = columns

	return schema, true
}

// [string, string]
func parseArray(items chan item) (array []string, i item, ok bool) {
	parseinfo("parseArray")

	i = <-items
	if i.typ != itemBeginArray {
		parseError(i, "expected [, got", i.val)
		return nil, i, false
	}

	i = <-items
	if i.typ != itemIdentifier {
		parseError(i, "expected identifier, got", i.val)
		return nil, i, false
	}
	array = append(array, i.val)

	i = <-items
	for {
		switch i.typ {
		case itemEndArray:
			return array, i, true
		case itemArrayDelimter:
			i = <-items
			if i.typ != itemIdentifier {
				parseError(i, "expected identifier, got", i.val)
				return nil, i, false
			}
			array = append(array, i.val)
		}
		i = <-items
	}
	return nil, i, false
}

// output <name> <schema>
func parseOutput(p *parser) (next parsefn, ok bool) {
	parseinfo("parseOutput")

	i := p.nextItem()
	if i.typ != itemIdentifier {
		parseError(i, "expected itemIdentifier, got ", i)
		return nil, false
	}

	name := i.val

	schema, ok := parseSchema(p.items)
	if !ok {
		return nil, false
	}

	if _, ok := p.s.inputs[name]; ok {
		parseError(i, "input", name, "already exists")
		return nil, false
	}

	schema.source = i.source
	p.s.outputs[name] = schema

	return parseSeed, true
}

// table <name> <schema>
func parseTable(p *parser) (next parsefn, ok bool) {
	parseinfo("parseTable")

	i := p.nextItem()
	if i.typ != itemIdentifier {
		parseError(i, "expected itemIdentifier, got ", i)
		return nil, false
	}

	name := i.val

	schema, ok := parseSchema(p.items)
	if !ok {
		return nil, false
	}

	if _, ok := p.s.inputs[name]; ok {
		parseError(i, "parseTable: input", name, "already exists")
		return nil, false
	}

	schema.source = i.source
	p.s.tables[name] = schema

	return parseSeed, true
}

func parseRule(p *parser) (next parsefn, ok bool) {
	parseinfo("parseRule")

	// destination
	destination := p.i

	// operation
	operation := p.nextItem()
	switch operation.typ {
	case itemOperationInsert, itemOperationSet,
		itemOperationDelete, itemOperationUpdate:
			//no-op
	default:
		parseError(operation, "expected an operation, got ", operation)
		return nil, false
	}

	// expression
	expr := p.nextItem()
	if expr.typ != itemIdentifier {
		parseError(expr, "expected itemIdentifier, got ", expr)
		return nil, false
	}

	r := new(rule)
	r.value = fmt.Sprint(destination.val, " ", operation.val, " ", expr.val)
	r.supplies = append(r.supplies, destination.val)
	r.requires = append(r.requires, expr.val)
	r.source = destination.source
	p.s.rules = append(p.s.rules, r)

	return parseSeed, true
}
