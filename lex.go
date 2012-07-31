// lexer for seed language
// based on http://golang.org/src/pkg/text/template/parse/lex.go

package main

import (
	"fmt"
	// "log"
	"strings"
	"unicode"
	"unicode/utf8"
)

func lexinfo(args ...interface{}) {
	// log.Println(args...)
}

type source struct {
	name string
	line int
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
	itemIdentifier // keywords and names
	itemBeginArray
	itemEndArray
	itemArrayDelimter
	itemOperationInsert
	itemMethodDelimiter
	// keywords
	itemKeyword // used to deliniate keyword identifiers
	itemInput   // input keyword
	itemOutput  // output keyword
	itemTable   // table keyword
)

var key = map[string]itemType{
	"input":  itemInput,
	"output": itemOutput,
	"table":  itemTable,
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

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos], l.source()}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.pos], "\n")
}

// returns a source struct for the line we are on
func (l *lexer) source() source {
	return source{l.name, l.lineNumber()}
}

// error returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...), l.source()}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
	panic("not reached")
}

// run lexes the input by executing state functions
// until the state is nil.
func (l *lexer) run() {
	for state := lexToken; state != nil; {
		state = state(l)
	}
	// log.Println("lexer stopped running")
	l.emit(itemEOF)
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		// state: lexText,
		items: make(chan item),
	}
	return l
}

// state functions

func lexToken(l *lexer) stateFn {
	lexinfo("lexToken")
	switch r := l.next(); {
	case r == eof:
		return nil
	case r == '#':
		return lexComment
	case unicode.IsLetter(r):
		return lexIdentifier
	case isSpace(r):
		l.ignore()
		return lexToken
	case r == '[':
		l.emit(itemBeginArray)
		return lexToken
	case r == ']':
		l.emit(itemEndArray)
		return lexToken
	case r == '<':
		return lexOperation
	case r == '.':
		l.emit(itemMethodDelimiter)
		return lexToken
	default:
		return l.errorf("lexToken: unrecognized character: %#U", r)
	}
	return nil
}

func lexComment(l *lexer) stateFn {
	lexinfo("lexComment")
	i := strings.Index(l.input[l.pos:], "\n")
	l.pos += i + len("\n")
	l.ignore()
	return lexToken
}

// lexIdentifier scans an alphanumeric or field.
func lexIdentifier(l *lexer) stateFn {
	lexinfo("lexIdentifier")
Loop:
	for {
		switch r := l.next(); {
		case r == '_' || unicode.IsLetter(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("lexIdentifier: unexpected character %+U '%s", r, r)
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
	lexinfo("lexOperation")
	switch r := l.next(); {
	case r == '+':
		l.emit(itemOperationInsert)
	}
	return lexToken
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) {
		return true
	}
	switch r {
	case eof, ',', '[', ']', '#', '.':
		return true
	}
	return false
}
