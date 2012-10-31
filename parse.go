package main

import (
	"fmt"
	"path"
	"runtime"
)

type parsefn func(p *parser) parsefn

type parser struct {
	s        *service
	items    chan item
	i        item // the last item
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
	parseinfo()

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

	fatal(message)
}

func parse(name, input string) *service {
	p := &parser{}
	p.s = &service{collections: make(map[string]*collection)}
	p.s.source = source{name: name}

	l := newLexer(name, input)
	go l.run()

	p.items = l.items

	for state := parseSeed; state != nil; {
		state = state(p)
	}

	return p.s
}

func parseSeed(p *parser) parsefn {
	parseinfo()

	i := p.next()

	switch i.typ {
	case itemInput:
		return parseCollection
	case itemOutput:
		return parseCollection
	case itemTable:
		return parseCollection
	case itemIdentifier:
		return parseRule
	case itemEOF:
		return nil
	default:
		p.error("unexpected", i)
	}

	return nil
}

// (input|output|table|scratch) <name> <schema>
func parseCollection(p *parser) parsefn {
	parseinfo()

	var collectionType collectionType
	switch p.i.typ {
	case itemInput:
		collectionType = collectionInput
	case itemOutput:
		collectionType = collectionOutput
	case itemTable:
		collectionType = collectionTable
	default:
		p.error("expected input, output, table, or scratch; got ", p.i)
	}

	i := p.next()
	if i.typ != itemIdentifier {
		p.error("expected itemIdentifier, got ", i)
	}

	name := i.val

	collection := new(collection)
	collection.key = parseArray(p)

	i = p.next()
	if i.typ == itemKeyRelation {
		columns := parseArray(p)
		collection.data = columns
	} else {
		p.backup()
	}

	if _, ok := p.s.collections[name]; ok {
		p.error("collection", name, "already exists")
	}

	collection.source = i.source
	collection.ctype = collectionType
	p.s.collections[name] = collection

	return parseSeed
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
	switch i.typ {
	case itemIdentifier:
		array = append(array, i.val)
	case itemEndArray:
		return array
	default:
		p.error("expected identifier or ], got", i.val)
	}

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

// <id> <op> <expr>
func parseRule(p *parser) parsefn {
	parseinfo()

	// destination
	destination := p.i
	r := &rule{source: destination.source}
	r.supplies = destination.val

	// operation
	operation := p.next()
	switch operation.typ {
	case itemOperationInsert, itemOperationDelete, itemOperationUpdate:
		r.operation = operation.val
	default:
		p.error("expected an operation, got ", operation)
	}

	// expr
	if p.next().typ != itemBeginArray {
		p.error("expected '[', got", p.i)
	}

	// get the array contents
	for {
		column := parseQualifiedColumn(p)
		r.projection = append(r.projection, column)

		if p.next().typ != itemArrayDelimter {
			break
		}
	}

	// using p.i so the previous loop does not need to backup
	if p.i.typ != itemEndArray {
		p.error("expected ']', got", p.i)
	}

	// if there is no ':', then the statement is finished
	if p.next().typ == itemPredicateDelimiter {
		// get the predicates
		for {
			left := parseQualifiedColumn(p)

			if p.next().typ != itemKeyRelation {
				p.error("expected '=>', got", p.i)
			}

			right := parseQualifiedColumn(p)

			r.predicate = append(r.predicate, constraint{left: left, right: right})

			if p.next().typ != itemArrayDelimter {
				p.backup()
				break
			}
		}
	} else {
		p.backup()
	}

	// do or reduce blocks (optional)
	p.next()
	if p.i.typ == itemDo || p.i.typ == itemReduce {
		reduce := false
		if p.i.typ == itemReduce {
			reduce = true
		}

		r.block = p.i.val

		if p.next().typ != itemPipe {
			p.error("expected '|', got", p.i)
		}

		if p.next().typ != itemIdentifier {
			p.error("expected identifier, got", p.i)
		}
		r.block = fmt.Sprintf("%s |%s", r.block, p.i.val)

		// reduce has two arguments
		if reduce {
			if p.next().typ != itemArrayDelimter {
				p.error("expected ',', got", p.i)
			}

			if p.next().typ != itemIdentifier {
				p.error("expected identifier, got", p.i)
			}

			r.block = fmt.Sprintf("%s, %s", r.block, p.i.val)
		}

		if p.next().typ != itemPipe {
			p.error("expected '|', got", p.i)
		}

		if p.next().typ != itemRuby {
			p.error("expected ruby, got", p.i)
		}
		r.block = fmt.Sprintf("%s|\n\t%s\nend", r.block, p.i.val)

		if p.next().typ != itemEnd {
			p.error("expected 'end', got", p.i)
		}
	} else {
		p.backup()
	}

	p.s.rules = append(p.s.rules, r)
	return parseSeed
}

func parseQualifiedColumn(p *parser) qualifiedColumn {
	parseinfo()

	collection := p.next()
	if collection.typ != itemIdentifier {
		p.error("expected identifier, got", collection)
	}

	if p.next().typ != itemScopeDelimiter {
		p.error("expected '.', got", p.i)
	}

	column := p.next()
	if column.typ != itemIdentifier {
		p.error("expected identifier, got", column)
	}

	return qualifiedColumn{collection: collection.val, column: column.val}
}
