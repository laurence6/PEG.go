package peg

import (
	"fmt"
	"runtime"
)

type Parser struct {
	tokens []*Token
	n      int    // current position
	token  *Token // current token
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{
		tokens: tokens,
		n:      0,
		token:  tokens[0],
	}
}

func (p *Parser) Parse() {
}

func (p *Parser) advance() {
	if p.n+1 >= len(p.tokens) {
		panic("Parser goes too far")
	}

	p.n += 1
	p.token = p.tokens[p.n]
}

func (p *Parser) back(n int) {
	if p.n-n < 0 {
		panic("Parser goes back too much")
	}

	p.n -= n
	p.token = p.tokens[p.n]
	return
}

func (p *Parser) expect(tt TokenType) error {
	if p.token.Type == tt {
		return nil
	} else {
		return NewTokenTypeError(2, tt, p.token.Type)
	}
}

func (p *Parser) peek(tt TokenType) bool {
	if p.n < len(p.tokens)-1 && p.tokens[p.n+1].Type == tt {
		return true
	}
	return false
}

type TokenTypeError struct {
	caller string
	expect TokenType
	got    TokenType
}

func NewTokenTypeError(skipCaller int, expect, got TokenType) error {
	caller, _, _, _ := runtime.Caller(skipCaller)
	callerName := runtime.FuncForPC(caller).Name()

	return TokenTypeError{
		caller: callerName,
		expect: expect,
		got:    got,
	}
}

func (e TokenTypeError) Error() string {
	return fmt.Sprintf("%s expect %v, got %v", e.caller, e.expect, e.got)
}

func (e TokenTypeError) String() string {
	return e.Error()
}
