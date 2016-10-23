package peg

import "fmt"

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
	if isLetter(s.char) {
		literal := []rune{s.char}
		for {
			if !isLetter(s.peekChar()) {
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
