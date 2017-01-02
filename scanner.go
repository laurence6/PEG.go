package peg

import (
	"fmt"
	"io"
	"unicode/utf8"
)

type Scanner struct {
	reader io.Reader
	eof    bool
	buf    []byte
	start  int
	end    int

	line int
	col  int
	char rune // current character
	next rune // next character
}

func NewScanner(reader io.Reader, bufsize int) *Scanner {
	if bufsize < utf8.UTFMax {
		panic("Scanner: Buffer size smaller than max utf-8 char bytes number")
	}
	s := &Scanner{
		reader: reader,
		buf:    make([]byte, bufsize),

		line: 1,
		col:  1,
	}
	s.nextChar()
	return s
}

func (s *Scanner) Scan() (token Token) {
Next:
	s.nextChar()

	token.Pos.Line = s.line
	token.Pos.Col = s.col

	tt := s.skipSpace()
	if tt == EOF {
		token.Type = tt
		return
	}

	if isIdentFirstChar(s.char) {
		literal := []rune{s.char}
		for {
			if !isIdentContChar(s.peekChar()) {
				break
			}
			s.nextChar()
			literal = append(literal, s.char)
		}

		if tt := isKeyword(literal); tt != NONE {
			token.Type = tt
		} else {
			token.Type = IDENT
			token.Literal = literal
		}
		return
	}

	switch s.char {
	case '#':
		for {
			s.nextChar()
			if s.char == utf8.RuneError || isNewline(s.char) {
				goto Next
			}
		}
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
				literal = append(literal, unescape(s.char))
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
				literal = append(literal, unescape(s.char))
			} else {
				literal = append(literal, s.char)
			}
		}
	case '{':
		literal := []rune{}
		depth := 0
		for {
			s.nextChar()
			if s.char == '{' {
				depth++
			} else if s.char == '}' && (len(literal) == 0 || literal[len(literal)-1] != '\\') {
				if depth == 0 {
					break
				} else {
					depth--
				}
			}
			literal = append(literal, s.char)
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

func (s *Scanner) fillBuf() {
	if s.start > 0 {
		copy(s.buf, s.buf[s.start:s.end])
		s.end -= s.start
		s.start = 0
	}
	if s.end >= len(s.buf) {
		return
	}

	n, _ := s.reader.Read(s.buf[s.end:])
	s.end += n
	if n == 0 {
		s.eof = true
	}

	return
}

func (s *Scanner) nextChar() {
	if s.end-s.start < utf8.UTFMax && !s.eof {
		s.fillBuf()
	}

	var r rune
	var size int
	if s.start == s.end {
		r, size = utf8.RuneError, 0
	} else {
		r, size = rune(s.buf[s.start]), 1
		if s.start == s.end || r >= utf8.RuneSelf {
			r, size = utf8.DecodeRune(s.buf[s.start:s.end])
		}
	}

	if r == utf8.RuneError {
		if size == 1 {
			panic("Invalid utf-8")
		}
	}

	if isNewline(s.next) {
		s.line += 1
		s.col = 1
	} else {
		s.col += 1
	}

	s.char = s.next
	s.next = r

	s.start += size
}

func (s *Scanner) peekChar() rune {
	return s.next
}

func (s *Scanner) skipSpace() (tt TokenType) {
	for {
		switch {
		case isSpace(s.char) || isNewline(s.char):
		case s.char == utf8.RuneError:
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

func unescape(char rune) rune {
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
