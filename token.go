package peg

import "fmt"

type Pos struct {
	Line int
	Col  int
}

type Token struct {
	Pos Pos

	Type    TokenType
	Literal []rune
}

func (t Token) String() string {
	return fmt.Sprintf("%d:%d %v %q", t.Pos.Line, t.Pos.Col, t.Type, string(t.Literal))
}

type TokenType uint8

const (
	NONE TokenType = iota // none
	EOF                   // EOF

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

	// keyword
	PACKAGE // package
	IMPORT  // import
)

// isKeyword returns corresponding TokenType if literal is keyword or returns NONE
func isKeyword(literal []rune) TokenType {
	lit := string(literal)
	switch lit {
	case "package":
		return PACKAGE
	case "import":
		return IMPORT
	default:
		return NONE
	}
}

func (tt TokenType) String() string {
	switch tt {
	case NONE:
		return "NONE" // none
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

	case PACKAGE:
		return "package"
	case IMPORT:
		return "import"
	}
	return "Unknown"
}
