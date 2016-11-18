package peg

import "fmt"

type Token struct {
	Type    TokenType
	Literal []rune
}

func (t Token) String() string {
	return fmt.Sprintf("%v %q", t.Type, string(t.Literal))
}

type TokenType uint8

const (
	ERROR TokenType = iota // error
	EOF                    // EOF

	IDENT  // abc
	STRING // "abc"
	RANGE  // [abc]
	CODE   // {abc}
	DOT    // .

	ASSIGN // =
	COLON  // :

	LPAREN // (
	RPAREN // )

	QUESTION // ?
	PLUS     // +
	STAR     // *

	AND // &
	NOT // !

	SLASH // /
)

func (tt TokenType) String() string {
	switch tt {
	case ERROR:
		return "ERROR" // error
	case EOF:
		return "EOF" // EOF

	case IDENT:
		return "IDENT" // abc
	case STRING:
		return "STRING" // "abc"
	case RANGE:
		return "RANGE" // [abc]
	case CODE:
		return "CODE" // {abc}
	case DOT:
		return "."

	case ASSIGN:
		return "="
	case COLON:
		return ":"

	case LPAREN:
		return "("
	case RPAREN:
		return ")"

	case QUESTION:
		return "?"
	case PLUS:
		return "+"
	case STAR:
		return "*"

	case AND:
		return "&"
	case NOT:
		return "!"

	case SLASH:
		return "/"
	}
	return "Unknown"
}
