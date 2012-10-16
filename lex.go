// lexer for seed language
// based on http://golang.org/src/pkg/text/template/parse/lex.go

package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

// toggle on and off by commenting the first return statement
func lexinfo(args ...interface{}) {
	return
	info(args...)
}

type source struct {
	name   string
	line   int
	column int
}

func (s source) String() string {
	return fmt.Sprint(s.name, ":", s.line)
}

// item represents a token or text string returned from the scanner.
type item struct {
	typ    itemType
	val    string
	source source
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 30:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemIdentifier         // keywords and names
	itemBeginArray         // [
	itemEndArray           // ]
	itemArrayDelimter      // ,
	itemOperationInsert    // <+
	itemOperationSet       // <=
	itemOperationDelete    // <-
	itemOperationUpdate    // <+-
	itemKeyRelation        // =>
	itemScopeDelimiter     // .
	itemPredicateDelimiter // :
	// keywords
	itemKeyword // used to deliniate keyword identifiers
	itemInput   // input keyword
	itemOutput  // output keyword
	itemTable   // table keyword
	itemScratch // scratch keyword
)

func (typ itemType) String() string {
	switch typ {
	case itemError:
		return "itemError"
	case itemEOF:
		return "itemEOF"
	case itemIdentifier:
		return "itemIdentifier"
	case itemBeginArray:
		return "itemBeginArray"
	case itemEndArray:
		return "itemEndArray"
	case itemArrayDelimter:
		return "itemArrayDelimter"
	case itemOperationInsert:
		return "itemOperationInsert"
	case itemOperationSet:
		return "itemOperationSet"
	case itemOperationDelete:
		return "itemOperationDelete"
	case itemOperationUpdate:
		return "itemOperationUpdate"
	case itemKeyRelation:
		return "itemKeyRelation"
	case itemScopeDelimiter:
		return "itemScopeDelimiter"
	case itemPredicateDelimiter:
		return "itemPredicateDelimiter"
	case itemKeyword:
		return "itemKeyword"
	case itemInput:
		return "itemInput"
	case itemOutput:
		return "itemOutput"
	case itemTable:
		return "itemTable"
	case itemScratch:
		return "itemScratch"
	}
	panic("unsupported type")
}

var key = map[string]itemType{
	"input":   itemInput,
	"output":  itemOutput,
	"table":   itemTable,
	"scratch": itemScratch,
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
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
func (l *lexer) source() source {
	return source{l.name, l.lineNumber(), l.columnNumber()}
}

// error returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) {
	message := ""

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		name := path.Ext(runtime.FuncForPC(pc).Name())
		name = name[1:]
		file = path.Base(file)
		message = fmt.Sprintf("%s:%d: [%s] ", file, line, name)
	}

	source := l.source()
	message += fmt.Sprintf("%s:%d: ERROR: ", source, source.column)

	message += fmt.Sprintf(format, args...)

	log.Fatalln(message)
	os.Exit(1)
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
	default:
		l.errorf("unrecognized character: %#U", r)
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
		case r == '_' || unicode.IsLetter(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				l.errorf("unexpected character %+U '%c'", r, r)
			}
			switch {
			case key[word] > itemKeyword:
				l.emit(key[word])
			default:
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
			l.emit(itemOperationUpdate) // <+-
		} else {
			l.emit(itemOperationInsert) // <+
		}
	case r == '=':
		l.emit(itemOperationSet) // <=
	case r == '-':
		l.emit(itemOperationDelete) // <-
	default:
		l.errorf("unexpected operation: '%s'", l.input[l.start:l.pos])
	}

	return lexToken
}

// itemKeyRelation =>
func lexKeyRelation(l *lexer) stateFn {
	lexinfo()

	if !l.accept(">") {
		l.errorf("expected '>', got: '%s'", l.input[l.start:l.pos])
	}
	l.emit(itemKeyRelation) // =>

	return lexToken
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

func (l *lexer) atTerminator() bool {
	lexinfo()

	r := l.peek()
	if isSpace(r) {
		return true
	}
	switch r {
	case eof, ',', '[', ']', '#', '.', '*', '(', ')':
		return true
	}
	return false
}
