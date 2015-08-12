package main

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

func parseExpressions(strs []string) []expression {
	exprs := []expression{}
	for _, str := range strs {
		expr, err := parseExpression(str)
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("%s: OK", str)
		exprs = append(exprs, expr)
	}
	return exprs
}

func parseExpression(str string) (expression, error) {
	var (
		expr   expression
		invert bool
	)

	_, c := lex("expression", str)
	for item := range c {
		switch item.itemType {
		case itemNot:
			invert = !invert

		case itemAll:
			expr.selector = selectAll

		case itemConnected:
			expr.selector = selectConnected

		case itemTouched:
			expr.selector = selectTouched

		case itemLike:
			item = <-c
			if item.itemType == itemError {
				return expression{}, fmt.Errorf(item.literal)
			}
			expr.selector = selectLike(item.literal)

		case itemWith:
			item = <-c
			switch item.itemType {
			case itemKeyValue:
				expr.selector = selectWith(item.literal)
			default:
				return expression{}, fmt.Errorf("bad WITH: %s", item.literal)
			}

		case itemRemove:
			expr.transformer = transformRemove

		case itemShowOnly:
			expr.transformer = transformShowOnly

		case itemMerge:
			expr.transformer = transformMerge

		case itemGroupBy:
			item = <-c
			if item.itemType == itemError {
				return expression{}, fmt.Errorf(item.literal)
			}
			expr.transformer = transformGroupBy(item.literal)

		default:
			return expression{}, fmt.Errorf("%s: %s", str, item.literal)
		}
	}
	if invert {
		expr.selector = selectNot(expr.selector)
	}
	return expr, nil
}

// Expression  = [NOT] Selector Transformer
// Selector    = ALL / CONNECTED / TOUCHED / LIKE {{ <regex> }} / WITH {{ <key> [= <value>] }}
// Transformer = REMOVE / SHOWONLY / MERGE / GROUPBY {{ <key>, ... }}

type itemType int

const (
	itemError itemType = iota
	itemNot
	itemAll
	itemConnected
	itemTouched
	itemRemove
	itemShowOnly
	itemMerge
	itemGroupBy
	itemLike
	itemWith
	itemRegex
	itemKeyValue
	itemKeyList
)

type stateFn func(*lexer) stateFn

type lexer struct {
	name  string // used only for errors
	input string // string being scanned
	start int    // start position of this item
	pos   int    // current position within the input
	width int    // width of last rune read
	items chan item
}

func lex(name, input string) (*lexer, <-chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l, l.items
}

func (l *lexer) run() {
	for state := lexExpression; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

const eof rune = -1

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

//func (l *lexer) ignore()    { l.start = l.pos }

func (l *lexer) backup() { l.pos -= l.width }

//func (l *lexer) peek() rune { r := l.next(); l.backup(); return r }

/*
// acceptOne consumes the next rune if it's from the valid set.
func (l *lexer) acceptOne(validSet string) bool {
	if strings.IndexRune(validSet, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}
*/

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(validSet string) {
	for strings.IndexRune(validSet, l.next()) >= 0 {
		// consume
	}
	l.backup()
}

func (l *lexer) eatWhitespace() {
	l.acceptRun(" \t\r\n")
}

// errorf terminates lexing with an error.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

type item struct {
	itemType itemType
	literal  string
}

const (
	keywordNot       = "NOT"
	keywordAll       = "ALL"
	keywordConnected = "CONNECTED"
	keywordTouched   = "TOUCHED"
	keywordLike      = "LIKE"
	keywordWith      = "WITH"
	keywordRemove    = "REMOVE"
	keywordShowOnly  = "SHOWONLY"
	keywordMerge     = "MERGE"
	keywordGroupBy   = "GROUPBY"
)

func lexExpression(l *lexer) stateFn {
	l.eatWhitespace()
	if strings.HasPrefix(l.input[l.pos:], keywordNot) {
		return lexNot
	}
	return lexSelector
}

func lexNot(l *lexer) stateFn {
	l.pos += len(keywordNot)
	l.emit(itemNot)
	return lexSelector
}

func lexSelector(l *lexer) stateFn {
	l.eatWhitespace()
	switch {
	case strings.HasPrefix(l.input[l.pos:], keywordAll):
		return lexAll
	case strings.HasPrefix(l.input[l.pos:], keywordConnected):
		return lexConnected
	case strings.HasPrefix(l.input[l.pos:], keywordTouched):
		return lexTouched
	case strings.HasPrefix(l.input[l.pos:], keywordLike):
		return lexLike
	case strings.HasPrefix(l.input[l.pos:], keywordWith):
		return lexWith
	default:
		return l.errorf("bad selector")
	}
}

func lexAll(l *lexer) stateFn {
	l.pos += len(keywordAll)
	l.emit(itemAll)
	return lexTransformer
}

func lexConnected(l *lexer) stateFn {
	l.pos += len(keywordConnected)
	l.emit(itemConnected)
	return lexTransformer
}

func lexTouched(l *lexer) stateFn {
	l.pos += len(keywordTouched)
	l.emit(itemTouched)
	return lexTransformer
}

func lexLike(l *lexer) stateFn {
	l.pos += len(keywordLike)
	l.emit(itemLike)
	return lexRegex
}

func lexWith(l *lexer) stateFn {
	l.pos += len(keywordWith)
	l.emit(itemWith)
	return lexKeyValue
}

func lexRegex(l *lexer) stateFn {
	return lexMeta("regex", itemRegex, lexTransformer)
}

func lexKeyValue(l *lexer) stateFn {
	return lexMeta("key=value", itemKeyValue, lexTransformer)
}

func lexTransformer(l *lexer) stateFn {
	l.eatWhitespace()
	switch {
	case strings.HasPrefix(l.input[l.pos:], keywordRemove):
		return lexRemove
	case strings.HasPrefix(l.input[l.pos:], keywordShowOnly):
		return lexShowOnly
	case strings.HasPrefix(l.input[l.pos:], keywordMerge):
		return lexMerge
	case strings.HasPrefix(l.input[l.pos:], keywordGroupBy):
		return lexGroupBy
	default:
		return l.errorf("bad transformer at position %d: %s", l.pos, l.input[l.pos:])
	}
}

func lexRemove(l *lexer) stateFn {
	l.pos += len(keywordRemove)
	l.emit(itemRemove)
	return nil
}

func lexShowOnly(l *lexer) stateFn {
	l.pos += len(keywordShowOnly)
	l.emit(itemShowOnly)
	return nil
}

func lexMerge(l *lexer) stateFn {
	l.pos += len(keywordMerge)
	l.emit(itemMerge)
	return nil
}

func lexGroupBy(l *lexer) stateFn {
	l.pos += len(keywordGroupBy)
	l.emit(itemGroupBy)
	return lexKeyList
}

func lexKeyList(l *lexer) stateFn {
	return lexMeta("key list", itemKeyList, nil)
}

const (
	leftMeta  = "{{"
	rightMeta = "}}"
)

func lexMeta(what string, item itemType, next stateFn) stateFn {
	return func(l *lexer) stateFn {
		l.eatWhitespace()
		if !strings.HasPrefix(l.input[l.pos:], leftMeta) {
			return l.errorf("%s must begin with %s", what, leftMeta)
		}
		l.pos += len(leftMeta)
		l.start = l.pos
		for {
			if l.pos > len(l.input) {
				return l.errorf("%s must end with %s", what, rightMeta)
			}
			if strings.HasPrefix(l.input[l.pos:], rightMeta) {
				break
			}
			l.pos++
		}
		l.emit(item)
		l.pos += len(rightMeta)
		l.start = l.pos
		return next
	}
}
