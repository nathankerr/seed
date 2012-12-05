// lexer for seed language
// based on http://golang.org/src/pkg/text/template/parse/lex.go

package service

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ    itemType
	val    string
	source Source
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case len(i.val) > 30:
		return fmt.Sprintf("%s: %.10q...", i.typ.String(), i.val)
	}
	return fmt.Sprintf("%s: %q", i.typ.String(), i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemEOF        itemType = iota
	itemIdentifier          // keywords and names
	itemRuby                // ruby blocks (contents of do and reduce blocks)
	// special characters, add the beginning of these to atTerminator
	itemBeginArray         // [
	itemEndArray           // ]
	itemArrayDelimter      // ,
	itemOperation          // <+, <-, <+-
	itemKeyRelation        // =>
	itemScopeDelimiter     // .
	itemPredicateDelimiter // :
	itemPipe               // |
	// keywords, also need to be in key map
	itemInput  // input keyword
	itemOutput // output keyword
	itemTable  // table keyword
	itemMap    // start of do block
	itemReduce // start of reduce block
	itemEnd    // end of do or reduce block
)

var key = map[string]itemType{
	"input":  itemInput,
	"output": itemOutput,
	"table":  itemTable,
	"map":    itemMap,
	"reduce": itemReduce,
	"end":    itemEnd,
}

func (l *lexer) atTerminator() bool {
	lexinfo()

	r := l.peek()
	if isSpace(r) {
		return true
	}
	switch r {
	case eof, ',', '[', ']', '#', '.', '(', ')', ':', '|', '<':
		// # is a special case for comments, which are not passed to the parser
		return true
	}
	return false
}

var itemNames = map[itemType]string{
	itemEOF:                "itemEOF",
	itemIdentifier:         "itemIdentifier",
	itemRuby:               "itemRuby",
	itemBeginArray:         "itemBeginArray",
	itemEndArray:           "itemEndArray",
	itemArrayDelimter:      "itemArrayDelimter",
	itemOperation:          "itemOperation",
	itemKeyRelation:        "itemKeyRelation",
	itemScopeDelimiter:     "itemScopeDelimiter",
	itemPredicateDelimiter: "itemPredicateDelimiter",
	itemPipe:               "itemPipe",
	// keywords
	itemInput:  "itemInput",
	itemOutput: "itemOutput",
	itemTable:  "itemTable",
	itemMap:    "itemMap",
	itemReduce: "itemReduce",
	itemEnd:    "itemEnd",
}

func (typ itemType) String() string {
	str, ok := itemNames[typ]
	if !ok {
		panic(fmt.Sprintf("itemType.String: unknown item type, %d", typ))
	}
	return str
}

const eof = -1

// state functions are used to drive the state machine
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name  string    // the name of the input; used only for error reports.
	input string    // the string being scanned.
	state stateFn   // the next lexing function to enter.
	pos   int       // current position in the input.
	start int       // start position of this item.
	width int       // width of last rune read from input.
	items chan item // channel of scanned items.
}

// lex creates a new scanner for the input string.
func newLexer(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	return l
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width

	lexinfo(string(r))
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	lexinfo()

	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	lexinfo()

	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	lexinfo(t.String())

	l.items <- item{t, l.input[l.start:l.pos], l.source()}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	lexinfo()

	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	lexinfo(valid)

	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	lexinfo(valid)

	for strings.IndexRune(valid, l.next()) >= 0 {
		// no-op
	}
	l.backup()
}

// lineNumber reports which line we're on. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.pos], "\n")
}

func (l *lexer) columnNumber() int {
	pos := l.start
	if pos >= len(l.input) {
		pos--
	}

	for pos != 0 {
		if l.input[pos] == '\n' {
			break
		}
		pos--
	}

	return l.start - pos
}

// returns a source struct for the line we are on
func (l *lexer) source() Source {
	return Source{l.name, l.lineNumber(), l.columnNumber()}
}

// run lexes the input by executing state functions
// until the state is nil.
func (l *lexer) run() {
	for state := lexToken; state != nil; {
		state = state(l)
	}
	lexinfo("lexer stopped running")
	l.emit(itemEOF)
}

// state functions

func lexToken(l *lexer) stateFn {
	lexinfo()

	switch r := l.next(); {
	case r == eof:
		return nil
	case r == '#':
		return lexComment
	case unicode.IsLetter(r):
		return lexIdentifier
	case isSpace(r):
		l.ignore()
	case r == '[':
		l.emit(itemBeginArray)
	case r == ']':
		l.emit(itemEndArray)
	case r == '<':
		return lexOperation
	case r == '=':
		return lexKeyRelation
	case r == ',':
		l.emit(itemArrayDelimter)
	case r == '.':
		l.emit(itemScopeDelimiter)
	case r == ':':
		l.emit(itemPredicateDelimiter)
	case r == '|':
		return lexRuby
	default:
		fatalf("%s unrecognized character: %#U",
			l.source(), r)
	}

	return lexToken
}

func lexComment(l *lexer) stateFn {
	lexinfo()

	i := strings.Index(l.input[l.pos:], "\n")
	l.pos += i + len("\n")
	l.ignore()

	return lexToken
}

// lexIdentifier scans an alphanumeric or field.
func lexIdentifier(l *lexer) stateFn {
	lexinfo()

Loop:
	for {
		switch r := l.next(); {
		case r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				fatalf("%s unexpected character %+U '%c'",
					l.source(), r, r)
			}

			typ, ok := key[word]
			if ok {
				l.emit(typ)
			} else {
				l.emit(itemIdentifier)
			}
			break Loop
		}
	}
	return lexToken
}

func lexOperation(l *lexer) stateFn {
	lexinfo()

	switch r := l.next(); {
	case r == '+':
		if l.peek() == '-' {
			l.next()
			l.emit(itemOperation) // <+-
		} else {
			l.emit(itemOperation) // <+
		}
	case r == '-':
		l.emit(itemOperation) // <-
	default:
		fatalf("%s unexpected operation: '%s'",
			l.source(), l.input[l.start:l.pos])
	}

	return lexToken
}

// itemKeyRelation =>
func lexKeyRelation(l *lexer) stateFn {
	lexinfo()

	if !l.accept(">") {
		fatalf("%s expected '>', got: '%s'",
			l.source(), l.input[l.start:l.pos])
	}
	l.emit(itemKeyRelation) // =>

	return lexToken
}

// | args | ruby 'end'
func lexRuby(l *lexer) stateFn {
	lexinfo()

	l.emit(itemPipe)

	// identifiers and arraydelimiters
Loop:
	for {
		switch r := l.next(); {
		case unicode.IsLetter(r):
			lexIdentifier(l)
		case isSpace(r):
			l.ignore()
		case r == ',':
			l.emit(itemArrayDelimter)
		case r == '|':
			break Loop
		default:
			fatalf("%s expected identifier, ',', or '|'; got: '%s'",
				l.source(), l.input[l.start:l.pos])
		}
	}
	l.emit(itemPipe)

	// eat the space between the end pipe and the ruby code
	for isSpace(l.next()) {
		// no-op
	}
	l.backup()
	l.start = l.pos

	start := l.start
	endpos := strings.Index(l.input[start:], "end")
	dopos := strings.Index(l.input[start:], "do")
	for dopos != -1 && dopos < endpos {
		endpos = strings.Index(l.input[start:], "end")
		dopos = strings.Index(l.input[start:], "do")
	}
	l.pos = l.start + endpos

	// eat the space between the ruby code and 'end'
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	for !isSpace(r) {
		l.pos--
		r, _ = utf8.DecodeRuneInString(l.input[l.pos:])

	}

	l.emit(itemRuby)

	for isSpace(l.next()) {
		l.ignore()
	}

	return lexIdentifier
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	lexinfo(r)

	switch r {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}
