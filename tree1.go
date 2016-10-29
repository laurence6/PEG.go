package peg

type Grammar struct {
	Code     []rune
	RuleList []*Rule
}

func (p *Parser) grammar() ([]rune, *Grammar, ret) {
	return nil, nil, NewRet(0)
}

func (p *Parser) ruleList() ([]*Rule, ret) {
	return nil, NewRet(0)
}

type Rule struct {
	Name       string
	ChoiceExpr *ChoiceExpr
}

func (p *Parser) rule() (*Rule, ret) {
	return nil, NewRet(0)
}

type ChoiceExpr struct {
	ActionExprs []*ActionExpr
}

func (p *Parser) choiceExpr() (*ChoiceExpr, ret) {
	return nil, NewRet(0)
}

type ActionExpr struct {
	SeqExpr *SeqExpr
	Code    []rune
}

func (p *Parser) actionExpr() (*ActionExpr, ret) {
	return nil, NewRet(0)
}

type SeqExpr struct {
	LabeledExprs []*LabeledExpr
}

func (p *Parser) seqExpr() (*SeqExpr, ret) {
	return nil, NewRet(0)
}

type LabeledExpr struct {
	Label        string
	PrefixedExpr *PrefixedExpr
}

func (p *Parser) labeledExpr() (*LabeledExpr, ret) {
	return nil, NewRet(0)
}

func (p *Parser) label() (string, ret) {
	return "", NewRet(0)
}

type PrefixedExpr struct {
	PrefixOp     TokenType
	SuffixedExpr *SuffixedExpr
}

type SuffixedExpr struct {
	PrimaryExpr *PrimaryExpr
	SuffixOp    TokenType
}

type PrimaryExpr struct {
	Type       int // 0 == Matcher, 1 == ChoiceExpr
	Matcher    TokenType
	ChoiceExpr *ChoiceExpr
}

func (p *Parser) primaryExpr() (*PrimaryExpr, ret) {
	return nil, NewRet(0)
}

func (p *Parser) prefixOp() (TokenType, ret) {
	return 0, NewRet(0)
}

func (p *Parser) suffixOp() (TokenType, ret) {
	return 0, NewRet(0)
}

func (p *Parser) ident() (string, ret) {
	return "", NewRet(0)
}

func (p *Parser) code() ([]rune, ret) {
	return nil, NewRet(0)
}

func (p *Parser) string() (string, ret) {
	return "", NewRet(0)
}

type ret struct {
	n   int // consumed
	err error
}

func (r ret) OK() bool {
	return r.err == nil
}

func NewRet(v interface{}) ret {
	switch v.(type) {
	case nil:
		return ret{}
	case int:
		return ret{n: v.(int)}
	case error:
		return ret{err: v.(error)}
	}
	panic("NewRet only accept int or error")
}
