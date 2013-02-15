package seed

type parsefn func(p *parser) parsefn

type parser struct {
	s        *Seed
	items    chan item
	i        item // the last item
	backedup bool // indicates i should be used instead of getting a new item
	subset   bool // limit to the subset
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

func Parse(name, input string, subset bool) *Seed {
	p := &parser{subset: subset}
	p.s = &Seed{Collections: make(map[string]*Collection)}
	p.s.Source = Source{Name: name, Line: 1, Column: 1}

	l := newLexer(name, input, subset)
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
	case itemChannel:
		if p.subset {
			fatal("channels only available when not in the subset")
			return nil
		}
		return parseCollection
	case itemScratch:
		if p.subset {
			fatal("scratch collections only available when not in the subset")
			return nil
		}
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
	case itemChannel:
		collectionType = CollectionChannel
	case itemScratch:
		collectionType = CollectionScratch
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

	// get the array contents, can be QualifiedColumns or FunctionCalls
loop:
	for {
		switch p.next().typ {
		case itemIdentifier:
			column := parseQualifiedColumn(p)
			r.Projection = append(r.Projection, Expression{Value: column})
		case itemStartParen:
			// FunctionCall
			functionCall := parseFunctionCall(p)
			r.Projection = append(r.Projection, Expression{Value: functionCall})
		case itemArrayDelimter:
			// no-op
		case itemEndArray:
			break loop
		default:
			fatal("expected identifier, '(', or ']', got", p.i)
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
			p.next()
			left := parseQualifiedColumn(p)

			if p.next().typ != itemKeyRelation {
				fatal("expected '=>', got", p.i)
			}

			p.next()
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

	p.s.Rules = append(p.s.Rules, r)
	return parseSeed
}

func parseFunctionCall(p *parser) FunctionCall {
	parseinfo()

	functionName := p.next()
	if functionName.typ != itemIdentifier {
		fatal("expected identifier, got", functionName)
	}

	arguments := []QualifiedColumn{}
loop:
	for {
		switch p.next().typ {
		case itemIdentifier:
			column := parseQualifiedColumn(p)
			arguments = append(arguments, column)
		case itemEndParen:
			break loop
		default:
			fatal("expected identifier, got", p.i)
		}
	}

	return FunctionCall{
		Name: functionName.val,
		Arguments: arguments,
	}
}

func parseQualifiedColumn(p *parser) QualifiedColumn {
	parseinfo()

	collection := p.i
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
