package peg

import (
	"fmt"
	"runtime"
)

type parser struct {
	tokens []*Token
	n      int    // current position
	token  *Token // current token
}

type Tree struct {
	*Grammar
}

func GetTree(tokens []*Token) *Tree {
	p := &parser{
		tokens: tokens,
		n:      0,
		token:  tokens[0],
	}

	grammar, ret := p.grammar()
	if !ret.OK() {
		panic(ret.err.Error())
	}

	return &Tree{grammar}
}

type Grammar struct {
	Code     string
	RuleList []*Rule
}

func (p *parser) grammar() (*Grammar, ret) {
	grammar := &Grammar{}
	n := 0

	code, r := p.code()
	if r.OK() {
		n += r.n
		grammar.Code = string(code)
	}

	ruleList, r := p.ruleList()
	if r.OK() {
		n += r.n
		grammar.RuleList = ruleList
	} else {
		p.back(n)
		return nil, r
	}

	if err := p.expect(EOF); err == nil {
	} else {
		p.back(n)
		return nil, newRet(err)
	}

	return grammar, newRet(n)
}

func (p *parser) ruleList() ([]*Rule, ret) {
	var ruleList []*Rule
	n := 0

	rule, r := p.rule()
	if r.OK() {
		n += r.n
		ruleList = []*Rule{rule}
	} else {
		return nil, r
	}

	for {
		rule, r = p.rule()
		if r.OK() {
			n += r.n
			ruleList = append(ruleList, rule)
		} else {
			break
		}
	}

	return ruleList, newRet(n)
}

type Rule struct {
	Name       string
	ChoiceExpr *ChoiceExpr
}

func (p *parser) rule() (*Rule, ret) {
	rule := &Rule{}
	n := 0

	if id, r := p.ident(); r.OK() {
		n += r.n
		rule.Name = id
	} else {
		return nil, r
	}

	if err := p.expect(ASSIGN); err == nil {
		p.advance()
		n += 1
	} else {
		p.back(n)
		return nil, newRet(err)
	}

	if e, r := p.choiceExpr(); r.OK() {
		n += r.n
		rule.ChoiceExpr = e
	} else {
		p.back(n)
		return nil, r
	}

	return rule, newRet(n)
}

type ChoiceExpr struct {
	ActionExprs []*ActionExpr
}

func (p *parser) choiceExpr() (*ChoiceExpr, ret) {
	exp := &ChoiceExpr{
		ActionExprs: []*ActionExpr{},
	}
	n := 0

	e, r := p.actionExpr()
	if r.OK() {
		n += r.n
		exp.ActionExprs = append(exp.ActionExprs, e)
	} else {
		return nil, r
	}

	for {
		err := p.expect(SLASH)
		if err == nil {
			p.advance()
			n += 1
		} else {
			break
		}

		e, r = p.actionExpr()
		if r.OK() {
			n += r.n
			exp.ActionExprs = append(exp.ActionExprs, e)
		} else {
			p.back(1)
			n -= 1
			break
		}
	}

	return exp, newRet(n)
}

type ActionExpr struct {
	SeqExpr *SeqExpr
	Code    string
}

func (p *parser) actionExpr() (*ActionExpr, ret) {
	exp := &ActionExpr{}
	n := 0

	e, r := p.seqExpr()
	if r.OK() {
		n += r.n
		exp.SeqExpr = e
	} else {
		return nil, r
	}

	if code, r := p.code(); r.OK() {
		n += r.n
		exp.Code = string(code)
	}

	return exp, newRet(n)
}

type SeqExpr struct {
	LabeledExprs []*LabeledExpr
}

func (p *parser) seqExpr() (*SeqExpr, ret) {
	exp := &SeqExpr{
		LabeledExprs: []*LabeledExpr{},
	}
	n := 0

	e, r := p.labeledExpr()
	if r.OK() {
		n += r.n
		exp.LabeledExprs = append(exp.LabeledExprs, e)
	} else {
		return nil, r
	}

	for {
		e, r = p.labeledExpr()
		if r.OK() {
			n += r.n
			exp.LabeledExprs = append(exp.LabeledExprs, e)
		} else {
			break
		}
	}

	return exp, newRet(n)
}

type LabeledExpr struct {
	Label        string
	PrefixedExpr *PrefixedExpr
}

func (p *parser) labeledExpr() (*LabeledExpr, ret) {
	exp := &LabeledExpr{}
	n := 0

	label, r := p.label()
	if r.OK() {
		n += r.n
		exp.Label = label
	}

	e, r := p.prefixedExpr()
	if r.OK() {
		n += r.n
		exp.PrefixedExpr = e
	} else {
		p.back(n)
		return nil, r
	}

	return exp, newRet(n)
}

func (p *parser) label() (string, ret) {
	n := 0

	label, r := p.ident()
	if r.OK() {
		n += r.n
	} else {
		return "", r
	}

	if err := p.expect(COLON); err == nil {
		p.advance()
		n += 1
	} else {
		p.back(n)
		return "", newRet(err)
	}

	return label, newRet(n)
}

type PrefixedExpr struct {
	PrefixOp     TokenType
	SuffixedExpr *SuffixedExpr
}

func (p *parser) prefixedExpr() (*PrefixedExpr, ret) {
	exp := &PrefixedExpr{}
	n := 0

	op, r := p.prefixOp()
	if r.OK() {
		n += p.n
		exp.PrefixOp = op
	}

	e, r := p.suffixedExpr()
	if r.OK() {
		n += r.n
		exp.SuffixedExpr = e
	} else {
		p.back(n)
		return nil, r
	}

	return exp, newRet(n)
}

func (p *parser) prefixOp() (TokenType, ret) {
	var err error
	if err = p.expect(AND); err == nil {
		p.advance()
		return AND, newRet(1)
	} else if err = p.expect(NOT); err == nil {
		p.advance()
		return NOT, newRet(1)
	}
	return 0, newRet(err)
}

type SuffixedExpr struct {
	PrimaryExpr *PrimaryExpr
	SuffixOp    TokenType
}

func (p *parser) suffixedExpr() (*SuffixedExpr, ret) {
	exp := &SuffixedExpr{}
	n := 0

	e, r := p.primaryExpr()
	if r.OK() {
		n += r.n
		exp.PrimaryExpr = e
	} else {
		return nil, r
	}

	op, r := p.suffixOp()
	if r.OK() {
		n += r.n
		exp.SuffixOp = op
	}

	return exp, newRet(n)
}

func (p *parser) suffixOp() (TokenType, ret) {
	var err error
	if err = p.expect(QUESTION); err == nil {
		p.advance()
		return QUESTION, newRet(1)
	} else if err = p.expect(PLUS); err == nil {
		p.advance()
		return PLUS, newRet(1)
	} else if err = p.expect(STAR); err == nil {
		p.advance()
		return STAR, newRet(1)
	}
	return 0, newRet(err)
}

type PrimaryExpr struct {
	PrimaryExpr interface{} // *Matcher / string (rule) / *ChoiceExpr
}

func (p *parser) primaryExpr() (*PrimaryExpr, ret) {
	exp := &PrimaryExpr{}
	n := 0

	if err := p.expect(STRING); err == nil {
		exp.PrimaryExpr = getMatcherString(string(p.token.Literal))
		p.advance()
		n += 1
	} else if err = p.expect(RANGE); err == nil {
		exp.PrimaryExpr = getMatcherRange(p.token.Literal)
		p.advance()
		n += 1
	} else if err = p.expect(DOT); err == nil {
		exp.PrimaryExpr = getMatcherRange([]rune("[^]"))
		p.advance()
		n += 1
	} else if id, r := p.ruleRef(); r.OK() {
		n += r.n
		exp.PrimaryExpr = id
	} else if e, r := p.subChoiceExpr(); r.OK() {
		n += r.n
		exp.PrimaryExpr = e
	} else {
		return nil, newRet(newTokenTypeError(1, STRING, p.token))
	}

	return exp, newRet(n)
}

func (p *parser) ruleRef() (string, ret) {
	n := 0

	name, r := p.ident()
	if r.OK() {
		n += r.n
	} else {
		return "", r
	}

	if err := p.expect(ASSIGN); err != nil {
	} else {
		p.back(n)
		return "", newRet(newTokenTypeError(1, 1, p.token))
	}

	return name, newRet(n)
}

func (p *parser) subChoiceExpr() (*ChoiceExpr, ret) {
	n := 0

	if err := p.expect(LPAREN); err == nil {
		p.advance()
		n += 1
	} else {
		return nil, newRet(err)
	}

	exp, r := p.choiceExpr()
	if r.OK() {
		n += r.n
	} else {
		p.back(n)
		return nil, r
	}

	if err := p.expect(RPAREN); err == nil {
		p.advance()
		n += 1
	} else {
		p.back(n)
		return nil, newRet(err)
	}

	return exp, newRet(n)
}

func (p *parser) ident() (string, ret) {
	err := p.expect(IDENT)
	if err == nil {
		ident := string(p.token.Literal)
		p.advance()
		return ident, newRet(1)
	}
	return "", newRet(err)
}

func (p *parser) code() ([]rune, ret) {
	err := p.expect(CODE)
	if err == nil {
		code := p.token.Literal
		p.advance()
		return code, newRet(1)
	}
	return nil, newRet(err)
}

func (p *parser) string() (string, ret) {
	err := p.expect(STRING)
	if err == nil {
		str := string(p.token.Literal)
		p.advance()
		return str, newRet(1)
	}
	return "", newRet(err)
}

func (p *parser) advance() {
	if p.n+1 >= len(p.tokens) {
		panic("Parser goes too far")
	}

	p.n += 1
	p.token = p.tokens[p.n]
}

func (p *parser) back(n int) {
	if p.n-n < 0 {
		panic("Parser goes back too much")
	}

	p.n -= n
	p.token = p.tokens[p.n]
	return
}

func (p *parser) expect(tt TokenType) error {
	if p.token.Type == tt {
		return nil
	} else {
		return newTokenTypeError(2, tt, p.token)
	}
}

type Matcher struct {
	Matcher interface{} // string / *CharRange
}

func getMatcherString(s string) *Matcher {
	return &Matcher{Matcher: s}
}

func getMatcherRange(r []rune) *Matcher {
	return &Matcher{Matcher: getCharRange(r)}
}

type CharRange struct {
	Not   bool
	Chars []*Char
}

func getCharRange(r []rune) *CharRange {
	chars := []*Char{}
	not := false

	r = r[1 : len(r)-1] // remove '[' ']'

	if len(r) > 0 && r[0] == '^' {
		not = true
		r = r[1:]
	}

	max := len(r) - 1
	i := 0
	for i <= max {
		if i+2 <= max && r[i+1] == '-' {
			chars = append(
				chars,
				newCharRange(r[i], r[i+2]),
			)
			i += 3
		} else {
			chars = append(
				chars,
				newCharRangeSingle(r[i]),
			)
			i += 1
		}
	}
	return &CharRange{Not: not, Chars: chars}
}

type Char struct {
	Start rune
	End   rune
}

func newCharRangeSingle(start rune) *Char {
	return &Char{
		Start: start,
		End:   start,
	}
}

func newCharRange(start, end rune) *Char {
	return &Char{
		Start: start,
		End:   end,
	}
}

type ret struct {
	n   int // consumed
	err error
}

func (r ret) OK() bool {
	return r.err == nil
}

func newRet(v interface{}) ret {
	switch v.(type) {
	case int:
		return ret{n: v.(int)}
	case error:
		return ret{err: v.(error)}
	}
	panic("newRet only accept int or error")
}

type tokenTypeError struct {
	caller string
	expect TokenType
	got    *Token
}

func newTokenTypeError(skipCaller int, expect TokenType, got *Token) error {
	caller, _, _, _ := runtime.Caller(skipCaller)
	callerName := runtime.FuncForPC(caller).Name()

	return tokenTypeError{
		caller: callerName,
		expect: expect,
		got:    got,
	}
}

func (e tokenTypeError) Error() string {
	return fmt.Sprintf("%d:%d %s expect %v, got %v", e.got.Pos.Line, e.got.Pos.Col, e.caller, e.expect, e.got.Type)
}

func (e tokenTypeError) String() string {
	return e.Error()
}
