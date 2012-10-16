package main

import (
	"fmt"
	"log"
	"path"
	"runtime"
)

// toggle on and off by commenting the first return statement
func parseinfo(args ...interface{}) {
	return
	info(args...)
}

type parsefn func(p *parser) parsefn

type parser struct {
	s        *seed
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

func parse(name, input string) *seed {
	p := &parser{}
	p.s = newSeed()

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
	case itemScratch:
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

	var collectionType seedCollectionType
	switch p.i.typ {
	case itemInput:
		collectionType = seedInput
	case itemOutput:
		collectionType = seedOutput
	case itemTable:
		collectionType = seedTable
	case itemScratch:
		collectionType = seedScratch
	default:
		p.error("expected input, output, table, or scratch; got ", p.i)
	}

	i := p.next()
	if i.typ != itemIdentifier {
		p.error("expected itemIdentifier, got ", i)
	}

	name := i.val

	collection := new(table)
	collection.key = parseArray(p)

	i = p.next()
	if i.typ == itemKeyRelation {
		columns := parseArray(p)
		collection.columns = columns
	} else {
		p.backup()
	}

	if _, ok := p.s.collections[name]; ok {
		p.error("collection", name, "already exists")
	}

	collection.source = i.source
	collection.typ = collectionType
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
	r := newRule(destination.source)
	r.supplies = destination.val

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

	// expr
	if p.next().typ != itemBeginArray {
		p.error("expected '[', got", p.i)
	}

	// get the array contents
	for {
		column := parseQualifiedColumn(p)
		r.requires[column.collection] = true
		r.output = append(r.output, column)

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
			r.requires[left.collection] = true

			if p.next().typ != itemKeyRelation {
				p.error("expected '=>', got", p.i)
			}

			right := parseQualifiedColumn(p)
			r.requires[right.collection] = true

			r.predicates = append(r.predicates, predicate{left: left, right: right})

			if p.next().typ != itemArrayDelimter {
				p.backup()
				break
			}
		}
	} else {
		p.backup()
	}

	p.s.rules = append(p.s.rules, r)
	return parseSeed
}

// [<id>(, <id>)*](: <id> => <id>(, <id> => <id>)*)
func parseJoin(p *parser, r *rule) *rule {
	parseinfo()

	

	return r
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