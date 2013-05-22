package seed

// TODO:
// Source

import (
	"fmt"
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
	ResetBuffer func(string) string
}

func (p *yyParser) Parse(ruleId int) (err error) {
	if p.rules[ruleId]() {
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
		err = &unexpectedEOFError{after}
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
		/* 1 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionInput
		},
		/* 2 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionOutput
		},
		/* 3 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionTable
		},
		/* 4 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionChannel
		},
		/* 5 CollectionType */
		func(yytext string, _ int) {
			yy.collectionType = CollectionScratch
		},
		/* 6 IdentifierArray */
		func(yytext string, _ int) {
			yy.strings = []string{}
		},
		/* 7 IdentifierArray */
		func(yytext string, _ int) {
			yy.strings = append(yy.strings, yytext)
		},
		/* 8 IdentifierArray */
		func(yytext string, _ int) {
			yy.strings = append(yy.strings, yytext)
		},
		/* 9 Rule */
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
		/* 10 Operation */
		func(yytext string, _ int) {
			yy.string = yytext
		},
		/* 11 Projection */
		func(yytext string, _ int) {
			e := yyval[yyp-1]
			yy.expressions = []Expression{}
			yyval[yyp-1] = e
		},
		/* 12 Projection */
		func(yytext string, _ int) {
			e := yyval[yyp-1]
			yy.expressions = append(yy.expressions, e.expression)
			yyval[yyp-1] = e
		},
		/* 13 Projection */
		func(yytext string, _ int) {
			e := yyval[yyp-1]
			yy.expressions = append(yy.expressions, e.expression)
			yyval[yyp-1] = e
		},
		/* 14 QualifiedColumn */
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
		/* 15 MapFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.mapfunction = MapFunction{Name: n.string}
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 16 MapFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.mapfunction.Arguments = append(yy.mapfunction.Arguments, c.expression.Value.(QualifiedColumn))
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 17 MapFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.mapfunction.Arguments = append(yy.mapfunction.Arguments, c.expression.Value.(QualifiedColumn))
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 18 MapFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.expression.Value = yy.mapfunction
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 19 ReduceFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.reducefunction = ReduceFunction{Name: n.string}
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 20 ReduceFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.reducefunction.Arguments = append(yy.reducefunction.Arguments, c.expression.Value.(QualifiedColumn))
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 21 ReduceFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.reducefunction.Arguments = append(yy.reducefunction.Arguments, c.expression.Value.(QualifiedColumn))
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 22 ReduceFunction */
		func(yytext string, _ int) {
			n := yyval[yyp-1]
			c := yyval[yyp-2]
			yy.expression.Value = yy.reducefunction
			yyval[yyp-1] = n
			yyval[yyp-2] = c
		},
		/* 23 Predicate */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			yy.constraints = append(yy.constraints, c.constraint)
			yyval[yyp-1] = c
		},
		/* 24 Predicate */
		func(yytext string, _ int) {
			c := yyval[yyp-1]
			yy.constraints = append(yy.constraints, c.constraint)
			yyval[yyp-1] = c
		},
		/* 25 Constraint */
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
		/* 26 Identifier */
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
		yyPush = 27 + iota
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

	commit := func(thunkPosition0 int) bool {
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
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
		l1:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[ruleStatement]() {
					goto l2
				}
				goto l1
			l2:
				position, thunkPosition = position2, thunkPosition2
			}
			if !p.rules[ruleEof]() {
				goto l0
			}
			if !(commit(thunkPosition0)) {
				goto l0
			}
			return true
		l0:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 1 Statement <- (Comment / Collection / Rule) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position4, thunkPosition4 := position, thunkPosition
				if !p.rules[ruleComment]() {
					goto l5
				}
				goto l4
			l5:
				position, thunkPosition = position4, thunkPosition4
				if !p.rules[ruleCollection]() {
					goto l6
				}
				goto l4
			l6:
				position, thunkPosition = position4, thunkPosition4
				if !p.rules[ruleRule]() {
					goto l3
				}
			}
		l4:
			return true
		l3:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 2 Comment <- ('#' < (!'\n' .)* > '\n'*) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('#') {
				goto l7
			}
			begin = position
		l8:
			{
				position9, thunkPosition9 := position, thunkPosition
				{
					position10, thunkPosition10 := position, thunkPosition
					if !matchChar('\n') {
						goto l10
					}
					goto l9
				l10:
					position, thunkPosition = position10, thunkPosition10
				}
				if !matchDot() {
					goto l9
				}
				goto l8
			l9:
				position, thunkPosition = position9, thunkPosition9
			}
			end = position
		l11:
			{
				position12, thunkPosition12 := position, thunkPosition
				if !matchChar('\n') {
					goto l12
				}
				goto l11
			l12:
				position, thunkPosition = position12, thunkPosition12
			}
			return true
		l7:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 3 Collection <- (CollectionType Spaces* Identifier Spaces* IdentifierArray (Spaces* '=>' Spaces* IdentifierArray)? Spaces* {
			p.Collections[n.string] = &Collection{
				Type: t.collectionType,
				Key: k.strings,
				Data: d.strings,
			}
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 4)
			if !p.rules[ruleCollectionType]() {
				goto l13
			}
			doarg(yySet, -1)
		l14:
			{
				position15, thunkPosition15 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l15
				}
				goto l14
			l15:
				position, thunkPosition = position15, thunkPosition15
			}
			if !p.rules[ruleIdentifier]() {
				goto l13
			}
			doarg(yySet, -2)
		l16:
			{
				position17, thunkPosition17 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l17
				}
				goto l16
			l17:
				position, thunkPosition = position17, thunkPosition17
			}
			if !p.rules[ruleIdentifierArray]() {
				goto l13
			}
			doarg(yySet, -3)
			{
				position18, thunkPosition18 := position, thunkPosition
			l20:
				{
					position21, thunkPosition21 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto l21
					}
					goto l20
				l21:
					position, thunkPosition = position21, thunkPosition21
				}
				if !matchString("=>") {
					goto l18
				}
			l22:
				{
					position23, thunkPosition23 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto l23
					}
					goto l22
				l23:
					position, thunkPosition = position23, thunkPosition23
				}
				if !p.rules[ruleIdentifierArray]() {
					goto l18
				}
				doarg(yySet, -4)
				goto l19
			l18:
				position, thunkPosition = position18, thunkPosition18
			}
		l19:
		l24:
			{
				position25, thunkPosition25 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l25
				}
				goto l24
			l25:
				position, thunkPosition = position25, thunkPosition25
			}
			do(0)
			doarg(yyPop, 4)
			return true
		l13:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 4 CollectionType <- (('input' { yy.collectionType = CollectionInput }) / ('output' { yy.collectionType = CollectionOutput }) / ('table' { yy.collectionType = CollectionTable }) / ('channel' { yy.collectionType = CollectionChannel }) / ('scratch' { yy.collectionType = CollectionScratch })) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position27, thunkPosition27 := position, thunkPosition
				if !matchString("input") {
					goto l28
				}
				do(1)
				goto l27
			l28:
				position, thunkPosition = position27, thunkPosition27
				if !matchString("output") {
					goto l29
				}
				do(2)
				goto l27
			l29:
				position, thunkPosition = position27, thunkPosition27
				if !matchString("table") {
					goto l30
				}
				do(3)
				goto l27
			l30:
				position, thunkPosition = position27, thunkPosition27
				if !matchString("channel") {
					goto l31
				}
				do(4)
				goto l27
			l31:
				position, thunkPosition = position27, thunkPosition27
				if !matchString("scratch") {
					goto l26
				}
				do(5)
			}
		l27:
			return true
		l26:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 5 IdentifierArray <- ('[' { yy.strings = []string{} } Spaces* Identifier { yy.strings = append(yy.strings, yytext) } (',' Spaces* Identifier { yy.strings = append(yy.strings, yytext) })* Spaces* ']') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('[') {
				goto l32
			}
			do(6)
		l33:
			{
				position34, thunkPosition34 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l34
				}
				goto l33
			l34:
				position, thunkPosition = position34, thunkPosition34
			}
			if !p.rules[ruleIdentifier]() {
				goto l32
			}
			do(7)
		l35:
			{
				position36, thunkPosition36 := position, thunkPosition
				if !matchChar(',') {
					goto l36
				}
			l37:
				{
					position38, thunkPosition38 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto l38
					}
					goto l37
				l38:
					position, thunkPosition = position38, thunkPosition38
				}
				if !p.rules[ruleIdentifier]() {
					goto l36
				}
				do(8)
				goto l35
			l36:
				position, thunkPosition = position36, thunkPosition36
			}
		l39:
			{
				position40, thunkPosition40 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l40
				}
				goto l39
			l40:
				position, thunkPosition = position40, thunkPosition40
			}
			if !matchChar(']') {
				goto l32
			}
			return true
		l32:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 6 Rule <- (Identifier Spaces* Operation Spaces* Projection (':' Spaces* Predicate)? Spaces* {
			p.Rules = append(p.Rules, &Rule{
				Supplies: c.string,
				Operation: o.string,
				Projection: proj.expressions,
				Predicate: pred.constraints,
			})
		}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 4)
			if !p.rules[ruleIdentifier]() {
				goto l41
			}
			doarg(yySet, -1)
		l42:
			{
				position43, thunkPosition43 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l43
				}
				goto l42
			l43:
				position, thunkPosition = position43, thunkPosition43
			}
			if !p.rules[ruleOperation]() {
				goto l41
			}
			doarg(yySet, -2)
		l44:
			{
				position45, thunkPosition45 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l45
				}
				goto l44
			l45:
				position, thunkPosition = position45, thunkPosition45
			}
			if !p.rules[ruleProjection]() {
				goto l41
			}
			doarg(yySet, -3)
			{
				position46, thunkPosition46 := position, thunkPosition
				if !matchChar(':') {
					goto l46
				}
			l48:
				{
					position49, thunkPosition49 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto l49
					}
					goto l48
				l49:
					position, thunkPosition = position49, thunkPosition49
				}
				if !p.rules[rulePredicate]() {
					goto l46
				}
				doarg(yySet, -4)
				goto l47
			l46:
				position, thunkPosition = position46, thunkPosition46
			}
		l47:
		l50:
			{
				position51, thunkPosition51 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l51
				}
				goto l50
			l51:
				position, thunkPosition = position51, thunkPosition51
			}
			do(9)
			doarg(yyPop, 4)
			return true
		l41:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 7 Operation <- (< ('<+-' / '<+' / '<-' / '<=' / '<~') > { yy.string = yytext }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			{
				position53, thunkPosition53 := position, thunkPosition
				if !matchString("<+-") {
					goto l54
				}
				goto l53
			l54:
				position, thunkPosition = position53, thunkPosition53
				if !matchString("<+") {
					goto l55
				}
				goto l53
			l55:
				position, thunkPosition = position53, thunkPosition53
				if !matchString("<-") {
					goto l56
				}
				goto l53
			l56:
				position, thunkPosition = position53, thunkPosition53
				if !matchString("<=") {
					goto l57
				}
				goto l53
			l57:
				position, thunkPosition = position53, thunkPosition53
				if !matchString("<~") {
					goto l52
				}
			}
		l53:
			end = position
			do(10)
			return true
		l52:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 8 Projection <- ('[' { yy.expressions = []Expression{} } Spaces* Expression { yy.expressions = append(yy.expressions, e.expression) } (',' Spaces* Expression { yy.expressions = append(yy.expressions, e.expression) })* Spaces* ']') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !matchChar('[') {
				goto l58
			}
			do(11)
		l59:
			{
				position60, thunkPosition60 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l60
				}
				goto l59
			l60:
				position, thunkPosition = position60, thunkPosition60
			}
			if !p.rules[ruleExpression]() {
				goto l58
			}
			doarg(yySet, -1)
			do(12)
		l61:
			{
				position62, thunkPosition62 := position, thunkPosition
				if !matchChar(',') {
					goto l62
				}
			l63:
				{
					position64, thunkPosition64 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto l64
					}
					goto l63
				l64:
					position, thunkPosition = position64, thunkPosition64
				}
				if !p.rules[ruleExpression]() {
					goto l62
				}
				doarg(yySet, -1)
				do(13)
				goto l61
			l62:
				position, thunkPosition = position62, thunkPosition62
			}
		l65:
			{
				position66, thunkPosition66 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l66
				}
				goto l65
			l66:
				position, thunkPosition = position66, thunkPosition66
			}
			if !matchChar(']') {
				goto l58
			}
			doarg(yyPop, 1)
			return true
		l58:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 9 Expression <- (MapFunction / ReduceFunction / QualifiedColumn) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position68, thunkPosition68 := position, thunkPosition
				if !p.rules[ruleMapFunction]() {
					goto l69
				}
				goto l68
			l69:
				position, thunkPosition = position68, thunkPosition68
				if !p.rules[ruleReduceFunction]() {
					goto l70
				}
				goto l68
			l70:
				position, thunkPosition = position68, thunkPosition68
				if !p.rules[ruleQualifiedColumn]() {
					goto l67
				}
			}
		l68:
			return true
		l67:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 10 QualifiedColumn <- (Identifier '.' Identifier { yy.expression.Value = QualifiedColumn{
			Collection: collection.string,
			Column: column.string,
		}}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleIdentifier]() {
				goto l71
			}
			doarg(yySet, -1)
			if !matchChar('.') {
				goto l71
			}
			if !p.rules[ruleIdentifier]() {
				goto l71
			}
			doarg(yySet, -2)
			do(14)
			doarg(yyPop, 2)
			return true
		l71:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 11 MapFunction <- ('(' Spaces* Identifier { yy.mapfunction = MapFunction{Name: n.string }} Spaces* QualifiedColumn { yy.mapfunction.Arguments = append(yy.mapfunction.Arguments, c.expression.Value.(QualifiedColumn)) } Spaces* (',' QualifiedColumn { yy.mapfunction.Arguments = append(yy.mapfunction.Arguments, c.expression.Value.(QualifiedColumn)) } Spaces*)* ')' Spaces* { yy.expression.Value = yy.mapfunction }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('(') {
				goto l72
			}
		l73:
			{
				position74, thunkPosition74 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l74
				}
				goto l73
			l74:
				position, thunkPosition = position74, thunkPosition74
			}
			if !p.rules[ruleIdentifier]() {
				goto l72
			}
			doarg(yySet, -1)
			do(15)
		l75:
			{
				position76, thunkPosition76 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l76
				}
				goto l75
			l76:
				position, thunkPosition = position76, thunkPosition76
			}
			if !p.rules[ruleQualifiedColumn]() {
				goto l72
			}
			doarg(yySet, -2)
			do(16)
		l77:
			{
				position78, thunkPosition78 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l78
				}
				goto l77
			l78:
				position, thunkPosition = position78, thunkPosition78
			}
		l79:
			{
				position80, thunkPosition80 := position, thunkPosition
				if !matchChar(',') {
					goto l80
				}
				if !p.rules[ruleQualifiedColumn]() {
					goto l80
				}
				doarg(yySet, -2)
				do(17)
			l81:
				{
					position82, thunkPosition82 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto l82
					}
					goto l81
				l82:
					position, thunkPosition = position82, thunkPosition82
				}
				goto l79
			l80:
				position, thunkPosition = position80, thunkPosition80
			}
			if !matchChar(')') {
				goto l72
			}
		l83:
			{
				position84, thunkPosition84 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l84
				}
				goto l83
			l84:
				position, thunkPosition = position84, thunkPosition84
			}
			do(18)
			doarg(yyPop, 2)
			return true
		l72:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 12 ReduceFunction <- ('{' Spaces* Identifier { yy.reducefunction = ReduceFunction{Name: n.string }} Spaces* QualifiedColumn { yy.reducefunction.Arguments = append(yy.reducefunction.Arguments, c.expression.Value.(QualifiedColumn)) } Spaces* (',' QualifiedColumn { yy.reducefunction.Arguments = append(yy.reducefunction.Arguments, c.expression.Value.(QualifiedColumn)) } Spaces*)* '}' Spaces* { yy.expression.Value = yy.reducefunction }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !matchChar('{') {
				goto l85
			}
		l86:
			{
				position87, thunkPosition87 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l87
				}
				goto l86
			l87:
				position, thunkPosition = position87, thunkPosition87
			}
			if !p.rules[ruleIdentifier]() {
				goto l85
			}
			doarg(yySet, -1)
			do(19)
		l88:
			{
				position89, thunkPosition89 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l89
				}
				goto l88
			l89:
				position, thunkPosition = position89, thunkPosition89
			}
			if !p.rules[ruleQualifiedColumn]() {
				goto l85
			}
			doarg(yySet, -2)
			do(20)
		l90:
			{
				position91, thunkPosition91 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l91
				}
				goto l90
			l91:
				position, thunkPosition = position91, thunkPosition91
			}
		l92:
			{
				position93, thunkPosition93 := position, thunkPosition
				if !matchChar(',') {
					goto l93
				}
				if !p.rules[ruleQualifiedColumn]() {
					goto l93
				}
				doarg(yySet, -2)
				do(21)
			l94:
				{
					position95, thunkPosition95 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto l95
					}
					goto l94
				l95:
					position, thunkPosition = position95, thunkPosition95
				}
				goto l92
			l93:
				position, thunkPosition = position93, thunkPosition93
			}
			if !matchChar('}') {
				goto l85
			}
		l96:
			{
				position97, thunkPosition97 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l97
				}
				goto l96
			l97:
				position, thunkPosition = position97, thunkPosition97
			}
			do(22)
			doarg(yyPop, 2)
			return true
		l85:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 13 Predicate <- (Constraint { yy.constraints = append(yy.constraints, c.constraint) } Spaces* (',' Spaces* Constraint { yy.constraints = append(yy.constraints, c.constraint) } Spaces*)*) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 1)
			if !p.rules[ruleConstraint]() {
				goto l98
			}
			doarg(yySet, -1)
			do(23)
		l99:
			{
				position100, thunkPosition100 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l100
				}
				goto l99
			l100:
				position, thunkPosition = position100, thunkPosition100
			}
		l101:
			{
				position102, thunkPosition102 := position, thunkPosition
				if !matchChar(',') {
					goto l102
				}
			l103:
				{
					position104, thunkPosition104 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto l104
					}
					goto l103
				l104:
					position, thunkPosition = position104, thunkPosition104
				}
				if !p.rules[ruleConstraint]() {
					goto l102
				}
				doarg(yySet, -1)
				do(24)
			l105:
				{
					position106, thunkPosition106 := position, thunkPosition
					if !p.rules[ruleSpaces]() {
						goto l106
					}
					goto l105
				l106:
					position, thunkPosition = position106, thunkPosition106
				}
				goto l101
			l102:
				position, thunkPosition = position102, thunkPosition102
			}
			doarg(yyPop, 1)
			return true
		l98:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 14 Constraint <- (QualifiedColumn Spaces* '=>' Spaces* QualifiedColumn { yy.constraint = Constraint {
			Left: l.expression.Value.(QualifiedColumn),
			Right: r.expression.Value.(QualifiedColumn),
		}}) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			doarg(yyPush, 2)
			if !p.rules[ruleQualifiedColumn]() {
				goto l107
			}
			doarg(yySet, -1)
		l108:
			{
				position109, thunkPosition109 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l109
				}
				goto l108
			l109:
				position, thunkPosition = position109, thunkPosition109
			}
			if !matchString("=>") {
				goto l107
			}
		l110:
			{
				position111, thunkPosition111 := position, thunkPosition
				if !p.rules[ruleSpaces]() {
					goto l111
				}
				goto l110
			l111:
				position, thunkPosition = position111, thunkPosition111
			}
			if !p.rules[ruleQualifiedColumn]() {
				goto l107
			}
			doarg(yySet, -2)
			do(25)
			doarg(yyPop, 2)
			return true
		l107:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 15 Identifier <- (< [a-zA-Z] [-a-zA-Z0-9_]+ > { yy.string = yytext }) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !matchClass(0) {
				goto l112
			}
			if !matchClass(1) {
				goto l112
			}
		l113:
			{
				position114, thunkPosition114 := position, thunkPosition
				if !matchClass(1) {
					goto l114
				}
				goto l113
			l114:
				position, thunkPosition = position114, thunkPosition114
			}
			end = position
			do(26)
			return true
		l112:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 16 Spaces <- (' ' / '\t' / '\n' / '\r') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position116, thunkPosition116 := position, thunkPosition
				if !matchChar(' ') {
					goto l117
				}
				goto l116
			l117:
				position, thunkPosition = position116, thunkPosition116
				if !matchChar('\t') {
					goto l118
				}
				goto l116
			l118:
				position, thunkPosition = position116, thunkPosition116
				if !matchChar('\n') {
					goto l119
				}
				goto l116
			l119:
				position, thunkPosition = position116, thunkPosition116
				if !matchChar('\r') {
					goto l115
				}
			}
		l116:
			return true
		l115:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 17 Eof <- !. */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position121, thunkPosition121 := position, thunkPosition
				if !matchDot() {
					goto l121
				}
				goto l120
			l121:
				position, thunkPosition = position121, thunkPosition121
			}
			return true
		l120:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
	}
}
