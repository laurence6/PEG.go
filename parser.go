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
	Code     []rune
	RuleList []*Rule
}

func (p *parser) grammar() (*Grammar, ret) {
	grammar := &Grammar{}
	n := 0

	code, r := p.code()
	if r.OK() {
		n += r.n
		grammar.Code = code
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
	Code    []rune
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
		exp.Code = code
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
	Type int // 0 == Matcher, 1 == Rule, 2 == ChoiceExpr

	Matcher    *Token
	RuleName   string
	ChoiceExpr *ChoiceExpr
}

func (p *parser) primaryExpr() (*PrimaryExpr, ret) {
	exp := &PrimaryExpr{}
	n := 0

	if err := p.expect(STRING); err == nil {
		exp.Matcher = p.token
		p.advance()
		n += 1
	} else if err = p.expect(RANGE); err == nil {
		exp.Matcher = p.token
		p.advance()
		n += 1
	} else if err = p.expect(DOT); err == nil {
		exp.Matcher = p.token
		p.advance()
		n += 1
	} else if id, r := p.ruleRef(); r.OK() {
		n += r.n
		exp.Type = 1
		exp.RuleName = id
	} else if e, r := p.subChoiceExpr(); r.OK() {
		n += r.n
		exp.Type = 2
		exp.ChoiceExpr = e
	} else {
		return nil, newRet(newTokenTypeError(1, STRING, p.token.Type))
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
		return "", newRet(newTokenTypeError(1, 1, ASSIGN))
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
		return newTokenTypeError(2, tt, p.token.Type)
	}
}

func (p *parser) peek(tt TokenType) bool {
	if p.n < len(p.tokens)-1 && p.tokens[p.n+1].Type == tt {
		return true
	}
	return false
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
	got    TokenType
}

func newTokenTypeError(skipCaller int, expect, got TokenType) error {
	caller, _, _, _ := runtime.Caller(skipCaller)
	callerName := runtime.FuncForPC(caller).Name()

	return tokenTypeError{
		caller: callerName,
		expect: expect,
		got:    got,
	}
}

func (e tokenTypeError) Error() string {
	return fmt.Sprintf("%s expect %v, got %v", e.caller, e.expect, e.got)
}

func (e tokenTypeError) String() string {
	return e.Error()
}
