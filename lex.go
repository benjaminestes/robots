package robots

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type fieldtype int

const (
	itemError = fieldtype(iota)
	itemUserAgent
	itemDisallow
	itemAllow
	itemSitemap
)

const eof = -1

var fieldtypes = map[string]fieldtype{
	"user-agent": itemUserAgent,
	"disallow":   itemDisallow,
	"allow":      itemAllow,
	"sitemap":    itemSitemap,
}

type item struct {
	typ  fieldtype
	val  string
	line int
}

type lexer struct {
	typ   fieldtype
	input string
	start int
	pos   int
	width int
	items chan *item
}

func (l *lexer) nextItem() *item {
	return <-l.items
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += w
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) emit() {
	l.items <- &item{
		typ: l.typ,
		val: strings.TrimRightFunc(l.input[l.start:l.pos], unicode.IsSpace),
	}
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
}

// Unlike model lexer, does not terminate lexing.  Per
// https://developers.google.com/search/reference/robots_txt#abstract,
// simply accept lines that are valid and silently discard those that
// are not (even if received content is HTML).
func (l *lexer) errorf(format string, args ...interface{}) {
	l.items <- &item{
		typ: itemError,
		val: fmt.Sprintf(format, args...),
	}
}

func (l *lexer) run() {
	for fn := lexStart; fn != nil; fn = fn(l) {
	}
	close(l.items)
}

func lex(in string) []*item {
	l := &lexer{
		input: in,
		items: make(chan *item),
	}
	go l.run()
	items := []*item{}
	for item := l.nextItem(); item != nil; item = l.nextItem() {
		items = append(items, item)
	}
	return items
}

type lexfn func(*lexer) lexfn

func lexStart(l *lexer) lexfn {
	c := l.peek()
	switch {
	case c == eof:
		return nil
	case c == '#':
		return lexComment
	case unicode.IsSpace(c):
		skipLWS(l)
		return lexStart
	default:
		return lexField
	}
}

func lexField(l *lexer) lexfn {
	for field, typ := range fieldtypes {
		if len(l.input[l.start:]) < len(field) {
			// Couldn't be a match.
			continue
		}
		if strings.EqualFold(field, l.input[l.start:l.start+len(field)]) {
			l.typ = typ
			l.pos += len(field)
			l.ignore()
			return lexSep
		}
	}
	// does this continue (as it should?)
	l.errorf("unexpected field type: %s", l.input[l.start:])
	return lexNextLine
}

func lexNextLine(l *lexer) lexfn {
	for c := l.next(); c != '\n' && c != eof; c = l.next() {
	}
	l.ignore()
	return lexStart
}

func lexSep(l *lexer) lexfn {
	skipLWS(l)
	if c := l.next(); c != ':' {
		l.errorf("expected separator betweeen field and value")
		return lexNextLine
	}
	skipLWS(l)
	return lexValue
}

func lexValue(l *lexer) lexfn {
	for c := l.next(); !isCTL(c) && c != '#' && c != eof; c = l.next() {
	}
	l.backup()
	l.emit()
	return lexComment
}

func lexComment(l *lexer) lexfn {
	more := skipLWS(l)
	if !more {
		// New line of input
		return lexStart
	}
	// We ran into something on the same logical line. What is it?
	if c := l.peek(); c == '#' {
		for c := l.next(); c != '\n' && c != eof; c = l.next() {
		}
		l.backup()
		l.ignore()
		return lexEOL
	}
	// The current line looked like it would continue, but it didn't.
	// There was no comment. Therefore it actually ended on a newline.
	// We should treat the next line as fresh input.
	return lexStart
}

func lexEOL(l *lexer) lexfn {
	c := l.next()
	if c == '\n' {
		l.ignore()
		return lexStart
	}
	if c == eof {
		l.ignore()
		return lexStart
	}
	l.errorf("expected EOL")
	return lexNextLine
}

// RFC 1945
// http://www.ietf.org/rfc/rfc1945.txt
func isCTL(r rune) bool {
	if r < 32 || r == 127 {
		return true
	}
	return false
}

// does this need to check for newline?
func skipLWS(l *lexer) (more bool) {
	afterEOL := false
	more = true
	for c := l.next(); ; c = l.next() {
		if c == '\n' {
			afterEOL = true
			continue
		}
		if afterEOL && !(c == ' ' || c == '\t') {
			more = false
			break
		}
		if !unicode.IsSpace(c) {
			break
		}
		afterEOL = false
	}
	// We've gone one rune too far.
	l.backup()
	l.ignore()
	return more
}
