package main

import (
	"fmt"
	"log"
	"path"
	"path/filepath"
	"runtime"
)

// toggle on and off by commenting the first return statement
func parseinfo(args ...interface{}) {
	return
	info := ""

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		basepath, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		sourcepath, err := filepath.Rel(basepath, file)
		if err != nil {
			panic(err)
		}
		info += fmt.Sprintf("%s:%d: ", sourcepath, line)

		name := path.Ext(runtime.FuncForPC(pc).Name())
		info += name[1:]
		if len(args) > 0 {
			info += ": "
		}
	}
	info += fmt.Sprintln(args...)
	
	log.Print(info)
}

type parsefn func(p *parser) (next parsefn, ok bool)

type parser struct {
	s     *seed
	items chan item
	i     item // the last item
	backedup bool // indicates that the last item should be used instead of getting a new one
}

func (p *parser) next() item {
	if p.backedup {
		p.backedup = false
	} else {
		p.i = <-p.items
	}
	parseinfo(p.i)
	return p.i
}

func (p *parser) backup() {
	parseinfo("backing up")
	p.backedup = true
}

func (p *parser) error(args ...interface{}) {
	message := ""

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		name := path.Ext(runtime.FuncForPC(pc).Name())
		name = name[1:]
		file = path.Base(file)
		message = fmt.Sprintf("%s:%d: [%s] ", file, line, name)
	}

	message += fmt.Sprintf("%s:%d: ERROR: ", p.i.source, p.i.source.column)
	message += fmt.Sprintln(args...)

	log.Fatal(message)
}

func parse(name, input string) (s *seed, ok bool) {
	p := &parser{}
	p.s = newSeed()

	l := newLexer(name, input)
	go l.run()

	p.items = l.items

	for state := parseSeed; state != nil; {
		state, ok = state(p)
		if !ok {
			return nil, false
		}
	}

	return p.s, true
}

func parseSeed(p *parser) (next parsefn, ok bool) {
	parseinfo()

	i := p.next()

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
		p.error("unexpected", i)
	}

	return nil, true
}

// input <name> <schema>
func parseInput(p *parser) (next parsefn, ok bool) {
	parseinfo()

	i := p.next()
	if i.typ != itemIdentifier {
		p.error("expected itemIdentifier, got ", i)
	}

	name := i.val

	schema := parseSchema(p)

	if _, ok := p.s.inputs[name]; ok {
		p.error("input", name, "already exists")
	}

	schema.source = i.source
	p.s.inputs[name] = schema

	return parseSeed, true
}

// [key] [columns]
func parseSchema(p *parser) *table {
	parseinfo()

	schema := new(table)

	schema.key = parseArray(p)

	i := p.next()
	if i.typ == itemKeyRelation {
		columns := parseArray(p)
	schema.columns = columns
	} else {
		p.backup()
	}

	return schema
}

// [string, string]
func parseArray(p *parser) []string {
	parseinfo()

	var array []string

	i := p.next()
	if i.typ != itemBeginArray {
		p.error("expected [, got", i.val)
	}

	i = p.next()
	if i.typ != itemIdentifier {
		p.error("expected identifier, got", i.val)
	}
	array = append(array, i.val)

	i = p.next()
	for {
		switch i.typ {
		case itemEndArray:
			return array
		case itemArrayDelimter:
			i = p.next()
			if i.typ != itemIdentifier {
				p.error("expected identifier, got", i.val)
			}
			array = append(array, i.val)
		}
		i = p.next()
	}
	return nil
}

// output <name> <schema>
func parseOutput(p *parser) (next parsefn, ok bool) {
	parseinfo()

	i := p.next()
	if i.typ != itemIdentifier {
		p.error("expected itemIdentifier, got ", i)
	}

	name := i.val

	schema := parseSchema(p)

	if _, ok := p.s.inputs[name]; ok {
		p.error("input", name, "already exists")
	}

	schema.source = i.source
	p.s.outputs[name] = schema

	return parseSeed, true
}

// table <name> <schema>
func parseTable(p *parser) (next parsefn, ok bool) {
	parseinfo()

	i := p.next()
	if i.typ != itemIdentifier {
		p.error("expected itemIdentifier, got ", i)
	}

	name := i.val
	if _, ok := p.s.inputs[name]; ok {
		p.error("parseTable: input", name, "already exists")
	}

	schema := parseSchema(p)
	schema.source = i.source
	p.s.tables[name] = schema

	return parseSeed, true
}

// <id> <op> <expr>
func parseRule(p *parser) (next parsefn, ok bool) {
	parseinfo()

	r := newRule()

	// destination
	destination := p.i
	r.source = destination.source
	r.value = fmt.Sprint(destination.val)
	r.supplies = append(r.supplies, destination.val)

	// operation
	operation := p.next()
	switch operation.typ {
	case itemOperationInsert:
		r.typ = ruleInsert
	case itemOperationSet:
		r.typ = ruleSet
	case itemOperationDelete:
		r.typ = ruleDelete
	case itemOperationUpdate:
		r.typ = ruleUpdate
	default:
		p.error("expected an operation, got ", operation)
	}
	r.value = fmt.Sprint(r.value, " ", operation.val, " ")

	// <id> | (<haspair>) 
	expr := p.next()
	switch expr.typ {
	case itemIdentifier:
		r.requires = append(r.requires, expr.val)
		r.value = fmt.Sprint(r.value, " ", expr.val)
	case itemBeginParen:
		parseHashPairs(p, r)
	default:
		p.error("expected identifier or (, got", expr)
	}

	// optional .
	i := p.next()
	if i.typ == itemMethodDelimiter {
		r.value += "."

		// <id> <block | doblock | args>
		i = p.next()
		if i.typ != itemIdentifier {
			p.error("expeced identifier, got", i)
		}
		r.value += i.val

		switch i = p.next(); i.typ {
		case itemDoBlock:
			r.value += fmt.Sprint(" ", i.val)
		default:
			p.error("expected arguments, do block, or brace block; got", i)
		}
	} else {
		p.backup()
	}

	p.s.rules = append(p.s.rules, r)
	return parseSeed, true
}

// ( <id> <* <id>>)
func parseHashPairs(p *parser, r *rule) {
	parseinfo()

	r.value += "("

	i := p.next()
	if i.typ != itemIdentifier {
		p.error("expected identifier, got", i)
	}
	r.requires = append(r.requires, i.val)
	r.value += i.val

	for {
		i = p.next()
		switch i.typ {
		case itemHashDelimiter:
			r.value += " * "
			i := p.next()
			if i.typ != itemIdentifier {
				p.error("expected identifier, got", i)
			}
			r.requires = append(r.requires, i.val)
			r.value += i.val
		case itemEndParen:
			r.value += ")"
			return
		default:
			p.error("expected hash delim or ), got", i)
		}
	}
}