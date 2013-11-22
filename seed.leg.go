package seed

// TODO:
// Source

import (
	"fmt"
	"io"
)

type yyinterface interface{}

type yystype struct {
	collectionType CollectionType
	string         string
	strings        []string
	expression     Expression
	expressions    []Expression
	constraint     Constraint
	constraints    []Constraint
	mapfunction    MapFunction
	reducefunction ReduceFunction
}

const (
	ruleSeed = iota
	ruleStatement
	ruleComment
	ruleCollection
	ruleCollectionType
	ruleIdentifierArray
	ruleRule
	ruleOperation
	ruleProjection
	ruleExpression
	ruleQualifiedColumn
	ruleMapFunction
	ruleReduceFunction
	rulePredicate
	ruleConstraint
	ruleIdentifier
	ruleSpaces
	ruleEof
)

type yyParser struct {
	*Seed
	Buffer      string
	Min, Max    int
	rules       [18]func() bool
	commit      func(int) bool
	ResetBuffer func(string) string
}

func (p *yyParser) Parse(ruleId int) (err error) {
	if p.rules[ruleId]() {
		// Make sure thunkPosition is 0 (there may be a yyPop action on the stack).
		p.commit(0)
		return
	}
	return p.parseErr()
}

type errPos struct {
	Line, Pos int
}

func (e *errPos) String() string {
	return fmt.Sprintf("%d:%d", e.Line, e.Pos)
}

type unexpectedCharError struct {
	After, At errPos
	Char      byte
}

func (e *unexpectedCharError) Error() string {
	return fmt.Sprintf("%v: unexpected character '%c'", &e.At, e.Char)
}

type unexpectedEOFError struct {
	After errPos
}

func (e *unexpectedEOFError) Error() string {
	return fmt.Sprintf("%v: unexpected end of file", &e.After)
}

func (p *yyParser) parseErr() (err error) {
	var pos, after errPos
	pos.Line = 1
	for i, c := range p.Buffer[0:] {
		if c == '\n' {
			pos.Line++
			pos.Pos = 0
		} else {
			pos.Pos++
		}
		if i == p.Min {
			if p.Min != p.Max {
				after = pos
			} else {
				break
			}
		} else if i == p.Max {
			break
		}
	}
	if p.Max >= len(p.Buffer) {
		if p.Min == p.Max {
			err = io.EOF
		} else {
			err = &unexpectedEOFError{after}
		}
	} else {
		err = &unexpectedCharError{after, pos, p.Buffer[p.Max]}
	}
	return
}

func (p *yyParser) Init() {
	var position int
	var yyp int
	var yy yystype
	var yyval = make([]yystype, 256)

	actions := [...]func(string, int){
		/* 0 Collection */
		func(yytext string, _ int) {
			t := yyval[yyp-1]
			n := yyval[yyp-2]
			k := yyval[yyp-3]
			d := yyval[yyp-4]
			yy = yystype{}
			yyval[yyp-1] = t
			yyval[yyp-2] = n
			yyval[yyp-3] = k
			yyval[yyp-4] = d
		},
		/* 1 Collection */
		func(yytext string, _ int) {
			t := yyval[yyp-1]
			n := yyval[yyp-2]
			k := yyval[yyp-3]
			d := yyval[yyp-4]

			p.Collections[n.string] = &Collection{
				Type: t.collectionType,
				Key:  k.strings,
				Data: d.strings,
			}

			yyval[yyp-1] = t
			yyval[yyp-2] = n
			yyval[yyp-3] = k
			yyval[yyp-4] = d
		},
		/* 2 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionInput
		},
		/* 3 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionOutput
		},
		/* 4 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionTable
		},
		/* 5 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionChannel
		},
		/* 6 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionScratch
		},
		/* 7 IdentifierArray */
		func(yytext string, _ int) {
			yy.strings = []string{}
		},
		/* 8 IdentifierArray */
		func(yytext string, _ int) {
			yy.strings = append(yy.strings, yytext)
		},
		/* 9 IdentifierArray */
		func(yytext string, _ int) {
			yy.strings = append(yy.strings, yytext)
		},
		/* 10 Rule */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			o := yyval[yyp-2]
			proj := yyval[yyp-3]
			pred := yyval[yyp-4]
			yy.constraints = []Constraint{}
			yyval[yyp-1] = c
			yyval[yyp-2] = o
			yyval[yyp-3] = proj
			yyval[yyp-4] = pred
		},
		/* 11 Rule */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			o := yyval[yyp-2]
			proj := yyval[yyp-3]
			pred := yyval[yyp-4]

			p.Rules = append(p.Rules, &Rule{
				Supplies:   c.string,
				Operation:  o.string,
				Projection: proj.expressions,
				Predicate:  pred.constraints,
			})

			yyval[yyp-1] = c
			yyval[yyp-2] = o
			yyval[yyp-3] = proj
			yyval[yyp-4] = pred
		},
		/* 12 Operation */
		func(yytext string, _ int) {
			yy.string = yytext
		},
		/* 13 Projection */
		func(yytext string, _ int) {
			e := yyval[yyp-1]
			yy.expressions = []Expression{}
			yyval[yyp-1] = e
		},
		/* 14 Projection */
		func(yytext string, _ int) {
			e := yyval[yyp-1]
			yy.expressions = append(yy.expressions, e.expression)
			yyval[yyp-1] = e
		},
		/* 15 Projection */
		func(yytext string, _ int) {
			e := yyval[yyp-1]
			yy.expressions = append(yy.expressions, e.expression)
			yyval[yyp-1] = e
		},
		/* 16 QualifiedColumn */
		func(yytext string, _ int) {
			collection := yyval[yyp-1]
			column := yyval[yyp-2]
			yy.expression.Value = QualifiedColumn{
				Collection: collection.string,
				Column:     column.string,
			}
			yyval[yyp-1] = collection
			yyval[yyp-2] = column
		},
		/* 17 MapFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.mapfunction = MapFunction{Name: n.string}
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 18 MapFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.mapfunction.Arguments = append(yy.mapfunction.Arguments, c.expression.Value.(QualifiedColumn))
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 19 MapFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.mapfunction.Arguments = append(yy.mapfunction.Arguments, c.expression.Value.(QualifiedColumn))
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 20 MapFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.expression.Value = yy.mapfunction
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 21 ReduceFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.reducefunction = ReduceFunction{Name: n.string}
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 22 ReduceFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.reducefunction.Arguments = append(yy.reducefunction.Arguments, c.expression.Value.(QualifiedColumn))
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 23 ReduceFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.reducefunction.Arguments = append(yy.reducefunction.Arguments, c.expression.Value.(QualifiedColumn))
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 24 ReduceFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.expression.Value = yy.reducefunction
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 25 Predicate */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			yy.constraints = append(yy.constraints, c.constraint)
			yyval[yyp-1] = c
		},
		/* 26 Predicate */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			yy.constraints = append(yy.constraints, c.constraint)
			yyval[yyp-1] = c
		},
		/* 27 Constraint */
		func(yytext string, _ int) {
			l := yyval[yyp-1]
			r := yyval[yyp-2]
			yy.constraint = Constraint{
				Left:  l.expression.Value.(QualifiedColumn),
				Right: r.expression.Value.(QualifiedColumn),
			}
			yyval[yyp-1] = l
			yyval[yyp-2] = r
		},
		/* 28 Identifier */
		func(yytext string, _ int) {
			yy.string = yytext
		},

		/* yyPush */
		func(_ string, count int) {
			yyp += count
			if yyp >= len(yyval) {
				s := make([]yystype, cap(yyval)+256)
				copy(s, yyval)
				yyval = s
			}
		},
		/* yyPop */
		func(_ string, count int) {
			yyp -= count
		},
		/* yySet */
		func(_ string, count int) {
			yyval[yyp+count] = yy
		},
	}
	const (
		yyPush = 29 + iota
		yyPop
		yySet
	)

	type thunk struct {
		action     uint8
		begin, end int
	}
	var thunkPosition, begin, end int
	thunks := make([]thunk, 32)
	doarg := func(action uint8, arg int) {
		if thunkPosition == len(thunks) {
			newThunks := make([]thunk, 2*len(thunks))
			copy(newThunks, thunks)
			thunks = newThunks
		}
		t := &thunks[thunkPosition]
		thunkPosition++
		t.action = action
		if arg != 0 {
			t.begin = arg // use begin to store an argument
		} else {
			t.begin = begin
		}
		t.end = end
	}
	do := func(action uint8) {
		doarg(action, 0)
	}

	p.ResetBuffer = func(s string) (old string) {
		if position < len(p.Buffer) {
			old = p.Buffer[position:]
		}
		p.Buffer = s
		thunkPosition = 0
		position = 0
		p.Min = 0
		p.Max = 0
		end = 0
		return
	}

	p.commit = func(thunkPosition0 int) bool {
		if thunkPosition0 == 0 {
			s := ""
			for _, t := range thunks[:thunkPosition] {
				b := t.begin
				if b >= 0 && b <= t.end {
					s = p.Buffer[b:t.end]
				}
				magic := b
				actions[t.action](s, magic)
			}
			p.Min = position
			thunkPosition = 0
			return true
		}
		return false
	}
	matchDot := func() bool {
		if position < len(p.Buffer) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	matchChar := func(c byte) bool {
		if (position < len(p.Buffer)) && (p.Buffer[position] == c) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	matchString := func(s string) bool {
		length := len(s)
		next := position + length
		if (next <= len(p.Buffer)) && p.Buffer[position] == s[0] && (p.Buffer[position:next] == s) {
			position = next
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	classes := [...][32]uint8{
		1: {0, 0, 0, 0, 0, 32, 255, 3, 254, 255, 255, 135, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		0: {0, 0, 0, 0, 0, 0, 0, 0, 254, 255, 255, 7, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	matchClass := func(class uint) bool {
		if (position < len(p.Buffer)) &&
			((classes[class][p.Buffer[position]>>3] & (1 << (p.Buffer[position] & 7))) != 0) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	p.rules = [...]func() bool{

		/* 0 Seed <- (Statement* Eof commit) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleStatement]() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleEof]() {
				goto ko
			}
			if !(p.commit(thunkPosition0)) {
				goto ko
			}
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 1 Statement <- (Comment / Collection / Rule) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleComment]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				position, thunkPosition = position1, thunkPosition1
				if !p.rules[ruleCollection]() {
					goto nextAlt3
				}
				goto ok
			nextAlt3:
				position, thunkPosition = position1, thunkPosition1
				if !p.rules[ruleRule]() {
					goto ko
				}
			}
		ok:
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 2 Comment <- ('#' < (!'\n' .)* > '\n'*) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('#') {
				goto ko
			}
			begin = position
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				{
					position2, thunkPosition2 := position, thunkPosition
					if !matchChar('\n') {
						goto ok
					}
					goto out
				ok:
					position, thunkPosition = position2, thunkPosition2
				}
				if !matchDot() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			end = position
		loop4:
			{
				position3, thunkPosition3 := position, thunkPosition
				if !matchChar('\n') {
					goto out5
				}
				goto loop4
			out5:
				position, thunkPosition = position3, thunkPosition3
			}
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 3 Collection <- ({ yy=yystype{} } CollectionType Spaces* Identifier Spaces* IdentifierArray (Spaces* '=>' Spaces* IdentifierArray)? Spaces* {
			p.Collections[n.string] = &Collection{
				Type: t.collectionType,
				Key: k.strings,
				Data: d.strings,
			}
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 4)
			do(0)
			if !p.rules[ruleCollectionType]() {
				goto ko
			}
			doarg(yySet, -1)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleIdentifier]() {
				goto ko
			}
			doarg(yySet, -2)
		loop3:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out4
				}
				goto loop3
			out4:
				position, thunkPosition = position2, thunkPosition2
			}
			if !p.rules[ruleIdentifierArray]() {
				goto ko
			}
			doarg(yySet, -3)
			{
				position3, thunkPosition3 := position, thunkPosition
			loop7:
				{
					position4, thunkPosition4 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out8
					}
					goto loop7
				out8:
					position, thunkPosition = position4, thunkPosition4
				}
				if !matchString("=>") {
					goto ko5
				}
			loop9:
				{
					position5, thunkPosition5 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out10
					}
					goto loop9
				out10:
					position, thunkPosition = position5, thunkPosition5
				}
				if !p.rules[ruleIdentifierArray]() {
					goto ko5
				}
				doarg(yySet, -4)
				goto ok
			ko5:
				position, thunkPosition = position3, thunkPosition3
			}
		ok:
		loop11:
			{
				position6, thunkPosition6 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out12
				}
				goto loop11
			out12:
				position, thunkPosition = position6, thunkPosition6
			}
			do(1)
			doarg(yyPop, 4)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 4 CollectionType <- (('input' { yy.collectionType = CollectionInput }) / ('output' { yy.collectionType = CollectionOutput }) / ('table' { yy.collectionType = CollectionTable }) / ('channel' { yy.collectionType = CollectionChannel }) / ('scratch' { yy.collectionType = CollectionScratch })) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1, thunkPosition1 := position, thunkPosition
				if !matchString("input") {
					goto nextAlt
				}
				do(2)
				goto ok
			nextAlt:
				position, thunkPosition = position1, thunkPosition1
				if !matchString("output") {
					goto nextAlt3
				}
				do(3)
				goto ok
			nextAlt3:
				position, thunkPosition = position1, thunkPosition1
				if !matchString("table") {
					goto nextAlt4
				}
				do(4)
				goto ok
			nextAlt4:
				position, thunkPosition = position1, thunkPosition1
				if !matchString("channel") {
					goto nextAlt5
				}
				do(5)
				goto ok
			nextAlt5:
				position, thunkPosition = position1, thunkPosition1
				if !matchString("scratch") {
					goto ko
				}
				do(6)
			}
		ok:
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 5 IdentifierArray <- ('[' { yy.strings = []string{} } Spaces* Identifier { yy.strings = append(yy.strings, yytext) } Spaces* (',' Spaces* Identifier { yy.strings = append(yy.strings, yytext) } Spaces*)* ']') */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('[') {
				goto ko
			}
			do(7)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleIdentifier]() {
				goto ko
			}
			do(8)
		loop3:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out4
				}
				goto loop3
			out4:
				position, thunkPosition = position2, thunkPosition2
			}
		loop5:
			{
				position3, thunkPosition3 := position, thunkPosition
				if !matchChar(',') {
					goto out6
				}
			loop7:
				{
					position4, thunkPosition4 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out8
					}
					goto loop7
				out8:
					position, thunkPosition = position4, thunkPosition4
				}
				if !p.rules[ruleIdentifier]() {
					goto out6
				}
				do(9)
			loop9:
				{
					position5, thunkPosition5 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out10
					}
					goto loop9
				out10:
					position, thunkPosition = position5, thunkPosition5
				}
				goto loop5
			out6:
				position, thunkPosition = position3, thunkPosition3
			}
			if !matchChar(']') {
				goto ko
			}
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 6 Rule <- ({yy.constraints = []Constraint{} } Identifier Spaces* Operation Spaces* Projection (':' Spaces* Predicate)? Spaces* {
			p.Rules = append(p.Rules, &Rule{
				Supplies: c.string,
				Operation: o.string,
				Projection: proj.expressions,
				Predicate: pred.constraints,
			})
		}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 4)
			do(10)
			if !p.rules[ruleIdentifier]() {
				goto ko
			}
			doarg(yySet, -1)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleOperation]() {
				goto ko
			}
			doarg(yySet, -2)
		loop3:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out4
				}
				goto loop3
			out4:
				position, thunkPosition = position2, thunkPosition2
			}
			if !p.rules[ruleProjection]() {
				goto ko
			}
			doarg(yySet, -3)
			{
				position3, thunkPosition3 := position, thunkPosition
				if !matchChar(':') {
					goto ko5
				}
			loop7:
				{
					position4, thunkPosition4 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out8
					}
					goto loop7
				out8:
					position, thunkPosition = position4, thunkPosition4
				}
				if !p.rules[rulePredicate]() {
					goto ko5
				}
				doarg(yySet, -4)
				goto ok
			ko5:
				position, thunkPosition = position3, thunkPosition3
			}
		ok:
		loop9:
			{
				position5, thunkPosition5 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out10
				}
				goto loop9
			out10:
				position, thunkPosition = position5, thunkPosition5
			}
			do(11)
			doarg(yyPop, 4)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 7 Operation <- (< ('<+-' / '<+' / '<-' / '<=' / '<~') > { yy.string = yytext }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			{
				position1, thunkPosition1 := position, thunkPosition
				if !matchString("<+-") {
					goto nextAlt
				}
				goto ok
			nextAlt:
				position, thunkPosition = position1, thunkPosition1
				if !matchString("<+") {
					goto nextAlt3
				}
				goto ok
			nextAlt3:
				position, thunkPosition = position1, thunkPosition1
				if !matchString("<-") {
					goto nextAlt4
				}
				goto ok
			nextAlt4:
				position, thunkPosition = position1, thunkPosition1
				if !matchString("<=") {
					goto nextAlt5
				}
				goto ok
			nextAlt5:
				position, thunkPosition = position1, thunkPosition1
				if !matchString("<~") {
					goto ko
				}
			}
		ok:
			end = position
			do(12)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 8 Projection <- ('[' { yy.expressions = []Expression{} } Spaces* Expression { yy.expressions = append(yy.expressions, e.expression) } (',' Spaces* Expression { yy.expressions = append(yy.expressions, e.expression) })* Spaces* ']') */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto ko
			}
			do(13)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleExpression]() {
				goto ko
			}
			doarg(yySet, -1)
			do(14)
		loop3:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !matchChar(',') {
					goto out4
				}
			loop5:
				{
					position3, thunkPosition3 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out6
					}
					goto loop5
				out6:
					position, thunkPosition = position3, thunkPosition3
				}
				if !p.rules[ruleExpression]() {
					goto out4
				}
				doarg(yySet, -1)
				do(15)
				goto loop3
			out4:
				position, thunkPosition = position2, thunkPosition2
			}
		loop7:
			{
				position4, thunkPosition4 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out8
				}
				goto loop7
			out8:
				position, thunkPosition = position4, thunkPosition4
			}
			if !matchChar(']') {
				goto ko
			}
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 9 Expression <- (MapFunction / ReduceFunction / QualifiedColumn) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleMapFunction]() {
					goto nextAlt
				}
				goto ok
			nextAlt:
				position, thunkPosition = position1, thunkPosition1
				if !p.rules[ruleReduceFunction]() {
					goto nextAlt3
				}
				goto ok
			nextAlt3:
				position, thunkPosition = position1, thunkPosition1
				if !p.rules[ruleQualifiedColumn]() {
					goto ko
				}
			}
		ok:
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 10 QualifiedColumn <- (Identifier '.' Identifier { yy.expression.Value = QualifiedColumn{
			Collection: collection.string,
			Column: column.string,
		}}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleIdentifier]() {
				goto ko
			}
			doarg(yySet, -1)
			if !matchChar('.') {
				goto ko
			}
			if !p.rules[ruleIdentifier]() {
				goto ko
			}
			doarg(yySet, -2)
			do(16)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 11 MapFunction <- ('(' Spaces* Identifier { yy.mapfunction = MapFunction{Name: n.string }} Spaces* QualifiedColumn { yy.mapfunction.Arguments = append(yy.mapfunction.Arguments, c.expression.Value.(QualifiedColumn)) } Spaces* (QualifiedColumn { yy.mapfunction.Arguments = append(yy.mapfunction.Arguments, c.expression.Value.(QualifiedColumn)) } Spaces*)* ')' Spaces* { yy.expression.Value = yy.mapfunction }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('(') {
				goto ko
			}
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleIdentifier]() {
				goto ko
			}
			doarg(yySet, -1)
			do(17)
		loop3:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out4
				}
				goto loop3
			out4:
				position, thunkPosition = position2, thunkPosition2
			}
			if !p.rules[ruleQualifiedColumn]() {
				goto ko
			}
			doarg(yySet, -2)
			do(18)
		loop5:
			{
				position3, thunkPosition3 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out6
				}
				goto loop5
			out6:
				position, thunkPosition = position3, thunkPosition3
			}
		loop7:
			{
				position4, thunkPosition4 := position, thunkPosition
				if !p.rules[ruleQualifiedColumn]() {
					goto out8
				}
				doarg(yySet, -2)
				do(19)
			loop9:
				{
					position5, thunkPosition5 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out10
					}
					goto loop9
				out10:
					position, thunkPosition = position5, thunkPosition5
				}
				goto loop7
			out8:
				position, thunkPosition = position4, thunkPosition4
			}
			if !matchChar(')') {
				goto ko
			}
		loop11:
			{
				position6, thunkPosition6 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out12
				}
				goto loop11
			out12:
				position, thunkPosition = position6, thunkPosition6
			}
			do(20)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 12 ReduceFunction <- ('{' Spaces* Identifier { yy.reducefunction = ReduceFunction{Name: n.string }} Spaces* QualifiedColumn { yy.reducefunction.Arguments = append(yy.reducefunction.Arguments, c.expression.Value.(QualifiedColumn)) } Spaces* (QualifiedColumn { yy.reducefunction.Arguments = append(yy.reducefunction.Arguments, c.expression.Value.(QualifiedColumn)) } Spaces*)* '}' Spaces* { yy.expression.Value = yy.reducefunction }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('{') {
				goto ko
			}
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !p.rules[ruleIdentifier]() {
				goto ko
			}
			doarg(yySet, -1)
			do(21)
		loop3:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out4
				}
				goto loop3
			out4:
				position, thunkPosition = position2, thunkPosition2
			}
			if !p.rules[ruleQualifiedColumn]() {
				goto ko
			}
			doarg(yySet, -2)
			do(22)
		loop5:
			{
				position3, thunkPosition3 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out6
				}
				goto loop5
			out6:
				position, thunkPosition = position3, thunkPosition3
			}
		loop7:
			{
				position4, thunkPosition4 := position, thunkPosition
				if !p.rules[ruleQualifiedColumn]() {
					goto out8
				}
				doarg(yySet, -2)
				do(23)
			loop9:
				{
					position5, thunkPosition5 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out10
					}
					goto loop9
				out10:
					position, thunkPosition = position5, thunkPosition5
				}
				goto loop7
			out8:
				position, thunkPosition = position4, thunkPosition4
			}
			if !matchChar('}') {
				goto ko
			}
		loop11:
			{
				position6, thunkPosition6 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out12
				}
				goto loop11
			out12:
				position, thunkPosition = position6, thunkPosition6
			}
			do(24)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 13 Predicate <- (Constraint { yy.constraints = append(yy.constraints, c.constraint) } Spaces* (',' Spaces* Constraint { yy.constraints = append(yy.constraints, c.constraint) } Spaces*)*) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleConstraint]() {
				goto ko
			}
			doarg(yySet, -1)
			do(25)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
		loop3:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !matchChar(',') {
					goto out4
				}
			loop5:
				{
					position3, thunkPosition3 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out6
					}
					goto loop5
				out6:
					position, thunkPosition = position3, thunkPosition3
				}
				if !p.rules[ruleConstraint]() {
					goto out4
				}
				doarg(yySet, -1)
				do(26)
			loop7:
				{
					position4, thunkPosition4 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto out8
					}
					goto loop7
				out8:
					position, thunkPosition = position4, thunkPosition4
				}
				goto loop3
			out4:
				position, thunkPosition = position2, thunkPosition2
			}
			doarg(yyPop, 1)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 14 Constraint <- (QualifiedColumn Spaces* '=>' Spaces* QualifiedColumn { yy.constraint = Constraint {
			Left: l.expression.Value.(QualifiedColumn),
			Right: r.expression.Value.(QualifiedColumn),
		}}) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleQualifiedColumn]() {
				goto ko
			}
			doarg(yySet, -1)
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			if !matchString("=>") {
				goto ko
			}
		loop3:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto out4
				}
				goto loop3
			out4:
				position, thunkPosition = position2, thunkPosition2
			}
			if !p.rules[ruleQualifiedColumn]() {
				goto ko
			}
			doarg(yySet, -2)
			do(27)
			doarg(yyPop, 2)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 15 Identifier <- (< [a-zA-Z] [-a-zA-Z0-9_]+ > { yy.string = yytext }) */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchClass(0) {
				goto ko
			}
			if !matchClass(1) {
				goto ko
			}
		loop:
			{
				position1, thunkPosition1 := position, thunkPosition
				if !matchClass(1) {
					goto out
				}
				goto loop
			out:
				position, thunkPosition = position1, thunkPosition1
			}
			end = position
			do(28)
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 16 Spaces <- (' ' / '\t' / '\n' / '\r') */
		func() (match bool) {
			position0, thunkPosition0 := position, thunkPosition
			{
				position1, thunkPosition1 := position, thunkPosition
				if !matchChar(' ') {
					goto nextAlt
				}
				goto ok
			nextAlt:
				position, thunkPosition = position1, thunkPosition1
				if !matchChar('\t') {
					goto nextAlt3
				}
				goto ok
			nextAlt3:
				position, thunkPosition = position1, thunkPosition1
				if !matchChar('\n') {
					goto nextAlt4
				}
				goto ok
			nextAlt4:
				position, thunkPosition = position1, thunkPosition1
				if !matchChar('\r') {
					goto ko
				}
			}
		ok:
			match = true
			return
		ko:
			position, thunkPosition = position0, thunkPosition0
			return
		},
		/* 17 Eof <- !. */
		func() (match bool) {
			{
				position1, thunkPosition1 := position, thunkPosition
				if !matchDot() {
					goto ok
				}
				return
			ok:
				position, thunkPosition = position1, thunkPosition1
			}
			match = true
			return
		},
	}
}
