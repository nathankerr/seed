package service

import (
	"fmt"
)

type parsefn func(p *parser) parsefn

type parser struct {
	s        *Service
	items    chan item
	i        item // the last item
	backedup bool // indicates i should be used instead of getting a new item
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

func Parse(name, input string) *Service {
	p := &parser{}
	p.s = &Service{Collections: make(map[string]*Collection)}
	p.s.Source = Source{Name: name, Line: 1, Column: 1}

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
		fatal("unexpected", i)
	}

	return nil
}

// (input|output|table|scratch) <name> <schema>
func parseCollection(p *parser) parsefn {
	parseinfo()

	var collectionType CollectionType
	switch p.i.typ {
	case itemInput:
		collectionType = CollectionInput
	case itemOutput:
		collectionType = CollectionOutput
	case itemTable:
		collectionType = CollectionTable
	default:
		fatal("expected input, output, table, or scratch; got ", p.i)
	}

	i := p.next()
	if i.typ != itemIdentifier {
		fatal("expected itemIdentifier, got ", i)
	}

	name := i.val

	collection := new(Collection)
	collection.Source = i.source
	collection.Key = parseArray(p)

	i = p.next()
	if i.typ == itemKeyRelation {
		columns := parseArray(p)
		collection.Data = columns
	} else {
		p.backup()
	}

	if _, ok := p.s.Collections[name]; ok {
		fatal("collection", name, "already exists")
	}

	collection.Type = collectionType
	p.s.Collections[name] = collection

	return parseSeed
}

// [string, string]
func parseArray(p *parser) []string {
	parseinfo()

	var array []string

	i := p.next()
	if i.typ != itemBeginArray {
		fatal("expected [, got", i.val)
	}

	i = p.next()
	switch i.typ {
	case itemIdentifier:
		array = append(array, i.val)
	case itemEndArray:
		return array
	default:
		fatal("expected identifier or ], got", i.val)
	}

	i = p.next()
	for {
		switch i.typ {
		case itemEndArray:
			return array
		case itemArrayDelimter:
			i = p.next()
			if i.typ != itemIdentifier {
				fatal("expected identifier, got", i.val)
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
	r := &Rule{Source: destination.source}
	r.Supplies = destination.val

	// operation
	operation := p.next()
	switch operation.typ {
	case itemOperation:
		r.Operation = operation.val
	default:
		fatal("expected an operation, got ", operation)
	}

	// expr
	if p.next().typ != itemBeginArray {
		fatal("expected '[', got", p.i)
	}

	// get the array contents
	for {
		column := parseQualifiedColumn(p)
		r.Projection = append(r.Projection, column)

		if p.next().typ != itemArrayDelimter {
			break
		}
	}

	// using p.i so the previous loop does not need to backup
	if p.i.typ != itemEndArray {
		fatal("expected ']', got", p.i)
	}

	// if there is no ':', then the statement is finished
	if p.next().typ == itemPredicateDelimiter {
		// get the predicates
		for {
			left := parseQualifiedColumn(p)

			if p.next().typ != itemKeyRelation {
				fatal("expected '=>', got", p.i)
			}

			right := parseQualifiedColumn(p)

			r.Predicate = append(r.Predicate,
				Constraint{Left: left, Right: right})

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
	if p.i.typ == itemMap || p.i.typ == itemReduce {
		reduce := false
		if p.i.typ == itemReduce {
			reduce = true
		}

		r.Block = p.i.val

		if p.next().typ != itemPipe {
			fatal("expected '|', got", p.i)
		}

		if p.next().typ != itemIdentifier {
			fatal("expected identifier, got", p.i)
		}
		r.Block = fmt.Sprintf("%s |%s", r.Block, p.i.val)

		// reduce has two arguments
		if reduce {
			if p.next().typ != itemArrayDelimter {
				fatal("expected ',', got", p.i)
			}

			if p.next().typ != itemIdentifier {
				fatal("expected identifier, got", p.i)
			}

			r.Block = fmt.Sprintf("%s, %s", r.Block, p.i.val)
		}

		if p.next().typ != itemPipe {
			fatal("expected '|', got", p.i)
		}

		if p.next().typ != itemRuby {
			fatal("expected ruby, got", p.i)
		}
		r.Block = fmt.Sprintf("%s|\n\t%s\nend", r.Block, p.i.val)

		if p.next().typ != itemEnd {
			fatal("expected 'end', got", p.i)
		}
	} else {
		p.backup()
	}

	p.s.Rules = append(p.s.Rules, r)
	return parseSeed
}

func parseQualifiedColumn(p *parser) QualifiedColumn {
	parseinfo()

	collection := p.next()
	if collection.typ != itemIdentifier {
		fatal("expected identifier, got", collection)
	}

	if p.next().typ != itemScopeDelimiter {
		fatal("expected '.', got", p.i)
	}

	column := p.next()
	if column.typ != itemIdentifier {
		fatal("expected identifier, got", column)
	}

	return QualifiedColumn{Collection: collection.val, Column: column.val}
}