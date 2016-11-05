package peg

import (
	"fmt"
	"unicode/utf8"
)

type Scanner struct {
	src  []rune // source code
	n    int    // current position
	char rune   // current character
}

func NewScanner(src []byte) *Scanner {
	return &Scanner{
		src: byteToRune(src),
		n:   -1,
	}
}

func (s *Scanner) Next() (token Token) {
Next:
	s.nextChar()

	tt := s.skipSpace()
	if tt == EOF {
		token.Type = tt
		return
	}

	// ident
	if isIdentFirstChar(s.char) {
		literal := []rune{s.char}
		for {
			if !isIdentContChar(s.peekChar()) {
				break
			}
			s.nextChar()
			literal = append(literal, s.char)
		}
		token.Literal = literal
		token.Type = IDENT
		return
	}

	switch s.char {
	case '#':
		for {
			s.nextChar()
			if s.char == '\x00' || isNewline(s.char) {
				goto Next
			}
		}
	// string
	case '"':
		literal := []rune{}
		for {
			s.nextChar()
			if s.char == '"' {
				token.Type = STRING
				token.Literal = literal
				return
			} else if s.char == '\\' {
				s.nextChar()
				literal = append(literal, escape(s.char))
			} else {
				literal = append(literal, s.char)
			}
		}
	case '[':
		literal := []rune{s.char}
		for {
			s.nextChar()
			if s.char == ']' {
				literal = append(literal, ']')
				token.Type = RANGE
				token.Literal = literal
				return
			} else if s.char == '\\' {
				s.nextChar()
				literal = append(literal, escape(s.char))
			} else {
				literal = append(literal, s.char)
			}
		}
	// FIXME: } appears in string
	case '{':
		literal := []rune{s.char}
		depth := 0
		for {
			s.nextChar()
			literal = append(literal, s.char)
			if s.char == '{' {
				depth++
			} else if s.char == '}' && (len(literal) == 2 || literal[len(literal)-2] != '\\') {
				if depth == 0 {
					break
				} else {
					depth--
				}
			}
		}
		token.Type = CODE
		token.Literal = literal
	case '.':
		token.Type = DOT
	case '=':
		token.Type = ASSIGN
	case ':':
		token.Type = COLON
	case '(':
		token.Type = LPAREN
	case ')':
		token.Type = RPAREN
	case '?':
		token.Type = QUESTION
	case '+':
		token.Type = PLUS
	case '*':
		token.Type = STAR
	case '&':
		token.Type = AND
	case '!':
		token.Type = NOT
	case '/':
		token.Type = SLASH
	default:
		panic(fmt.Sprintf("%q invalid character", s.char))
	}
	return
}

func (s *Scanner) GetAllTokens() (tokens []*Token) {
	for {
		token := s.Next()
		tokens = append(tokens, &token)
		if token.Type == EOF {
			break
		}
	}
	return
}

func (s *Scanner) nextChar() {
	if s.n+1 >= len(s.src) {
		s.char = '\x00'
		return
	}

	s.n++
	s.char = s.src[s.n]
}

func (s *Scanner) peekChar() rune {
	if s.n >= len(s.src)-1 {
		return '\x00'
	}

	return s.src[s.n+1]
}

func (s *Scanner) skipSpace() (tt TokenType) {
	for {
		switch {
		case isSpace(s.char) || isNewline(s.char):
		case s.char == '\x00':
			tt = EOF
			return
		default:
			return
		}
		s.nextChar()
	}
}

func isIdentFirstChar(char rune) bool {
	if isLetter(char) || char == '_' {
		return true
	}
	return false
}

func isIdentContChar(char rune) bool {
	if isLetter(char) || isDigit(char) || char == '_' {
		return true
	}
	return false
}

func isNewline(char rune) bool {
	return char == '\n' || char == '\r'
}

func isSpace(char rune) bool {
	return char == ' ' || char == '\t'
}

func isLetter(char rune) bool {
	return ('A' <= char && char <= 'Z') || ('a' <= char && char <= 'z')
}

func isDigit(char rune) bool {
	return '0' <= char && char <= '9'
}

func lenRune(r rune) int {
	if 0x0 <= r && r <= 0x7f {
		return 1
	} else if 0x80 <= r && r <= 0x7ff {
		return 2
	} else if 0x800 <= r && r <= 0xffff {
		return 3
	} else {
		return 4
	}
}

func byteToRune(b []byte) []rune {
	runes := []rune{}
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		runes = append(runes, r)
		b = b[size:]
	}
	return runes
}

// TODO finish escape
func escape(char rune) rune {
	switch char {
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	}
	return char
}
