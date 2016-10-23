package peg

type Token struct {
	Type    TokenType
	Literal []rune
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
		return "DOT" // .

	case ASSIGN:
		return "ASSIGN" // =
	case COLON:
		return "COLON" // :

	case LPAREN:
		return "LPAREN" // (
	case RPAREN:
		return "RPAREN" // )

	case QUESTION:
		return "QUESTION" // ?
	case PLUS:
		return "PLUS" // +
	case STAR:
		return "STAR" // *

	case AND:
		return "AND" // &
	case NOT:
		return "NOT" // !

	case SLASH:
		return "SLASH" // /
	}
	return "Unknown"
}
