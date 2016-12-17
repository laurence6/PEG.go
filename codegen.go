package peg

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

const header = `
var pegErr = errors.New("PEG ERROR")

func main() {
	src, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, "Input:----\n"+ string(src)+ "\n----------")
	fmt.Println(Parse([]rune(string(src))))
}

func Parse(src []rune) (interface{}, error) {
	p := parser{src, 0}
	return p.rule_%s()
}

type parser struct {
	src []rune
	n   int
}

func (__p *parser) advance(n int) {
	__p.n += n
}

func (__p *parser) backTo(n int) {
	__p.n = n
}

func (__p *parser) expectDot() string {
	if __p.n < len(__p.src) {
		return string(__p.src[__p.n])
	}
	return ""
}

func (__p *parser) expectString(str string, l int) bool {
	if __p.n + l <= len(__p.src) && str == string(__p.src[__p.n:__p.n+l]) {
		return true
	}
	return false
}

func (__p *parser) expectChar(chars ...rune) string {
	if __p.n < len(__p.src) {
		c := __p.src[__p.n]
		for i := 0; i < len(chars); i += 2 {
			if chars[i] <= c && c <= chars[i+1] {
				return string(c)
			}
		}
	}
	return ""
}

func (__p *parser) expectCharNot(chars ...rune) string {
	if __p.n < len(__p.src) {
		c := __p.src[__p.n]
		for i := 0; i < len(chars); i += 2 {
			if c < chars[i] || chars[i+1] < c {
				return string(c)
			}
		}
	}
	return ""
}

func (__p *parser) zeroOrOne(pe func() (interface{}, error)) (interface{}, error) {
	if r, err := pe(); err == nil {
		return r, nil
	}
	return nil, nil
}

func (__p *parser) oneOrMore(pe func() (interface{}, error)) (interface{}, error) {
	var ret []interface{}
	if r, err := pe(); err == nil {
		ret = []interface{}{r}
	} else {
		return nil, pegErr
	}
	for {
		if r, err := pe(); err == nil {
			ret = append(ret, r)
		} else {
			break
		}
	}
	if len(ret) > 0 {
		return ret, nil
	} else {
		return nil, pegErr
	}
}

func (__p *parser) zeroOrMore(pe func() (interface{}, error)) (interface{}, error) {
	ret := []interface{}{}
	for {
		if r, err := pe(); err == nil {
			ret = append(ret, r)
		} else {
			break
		}
	}
	return ret, nil
}
`

func (tree *Tree) GenCode(out io.Writer) {
	fmt.Fprintf(out, "package %s\n", "main")

	fmt.Fprintf(out, header,
		tree.RuleList[0].Name)

	for _, r := range tree.RuleList {
		r.GenCode(out)
	}

	fmt.Fprint(out, tree.Grammar.Code)

	io.Copy(out, userCode)
}

func (r *Rule) GenCode(out io.Writer) {
	fmt.Fprint(out, "// Rule: ")
	r.Print(out)
	fmt.Fprintln(out, "")

	fmt.Fprintf(out, "func (__p *parser) rule_%s() (interface{}, error) {\n", r.Name)

	r.ChoiceExpr.GenCode(out)

	fmt.Fprintln(out, "return nil, pegErr")

	fmt.Fprintln(out, "}\n")
}

func (ce *ChoiceExpr) GenCode(out io.Writer) {
	fmt.Fprintln(out, "var __peg_n int")
	for _, ae := range ce.ActionExprs {
		fmt.Fprintln(out, "__peg_n = __p.n")
		fmt.Fprintf(out, "if __ae_ret, err := ")
		ae.GenCode(out)
		fmt.Fprintf(out,
			"; err == nil {\n"+
				"	return __ae_ret, nil\n"+
				"} else {\n"+
				"	__p.backTo(__peg_n)"+
				"}\n",
		)
	}
}

func (se *SeqExpr) hasLabel() bool {
	for _, le := range se.LabeledExprs {
		if le.Label != "" {
			return true
		}
	}
	return false
}

var userCodeN uint64 = 0
var userCode = &bytes.Buffer{}

func (ae *ActionExpr) GenCode(out io.Writer) {
	fmt.Fprint(out, "func() (interface{}, error) {\n")

	vars := []string{}
	hasLabel := ae.SeqExpr.hasLabel()
	for n, le := range ae.SeqExpr.LabeledExprs {
		varName := "_"
		if !hasLabel {
			varName = fmt.Sprintf("__peg_v%d", n)
		} else if le.Label != "" {
			varName = le.Label
		}
		if varName != "_" {
			vars = append(vars, varName)
			fmt.Fprintf(out, "var %s interface{}\n", varName)
		}

		fmt.Fprint(out, "if __pe_ret, err := ")

		le.PrefixedExpr.GenCode(out)

		valueVarOrEmpty := "__pe_ret"
		not := "="
		if le.PrefixedExpr.PrefixOp == AND || le.PrefixedExpr.PrefixOp == NOT {
			valueVarOrEmpty = "nil"
		}
		if le.PrefixedExpr.PrefixOp == NOT {
			not = "!"
		}
		fmt.Fprintf(out,
			"; err %s= nil {\n"+
				"	%s = %s\n"+
				"} else {\n"+
				"	return nil, pegErr\n"+
				"}\n",
			not,
			varName,
			valueVarOrEmpty,
		)
	}

	if ae.Code != "" {
		var paramsDef string
		var paramsCall string
		if hasLabel {
			paramsDef = fmt.Sprintf("%s interface{}", strings.Join(vars, ", "))
			paramsCall = strings.Join(vars, ", ")
		} else {
			paramsDef = fmt.Sprintf("result [%d]interface{}", len(vars))
			paramsCall = fmt.Sprintf("[...]interface{}{%s}", strings.Join(vars, ", "))
		}

		userCode.WriteString(
			fmt.Sprintf(
				"func (__p *parser) ae_code_%d(%s) (ret interface{}) {\n"+
					"	%s\n"+
					"	return\n"+
					"}\n",
				userCodeN,
				paramsDef,
				ae.Code,
			),
		)

		fmt.Fprintf(out, "return __p.ae_code_%d(%s), nil\n",
			userCodeN,
			paramsCall,
		)

		userCodeN++
	} else {
		if len(vars) > 1 {
			fmt.Fprintf(out, "return [...]interface{}{%s}, nil\n", strings.Join(vars, ", "))
		} else {
			fmt.Fprintf(out, "return %s, nil\n", vars[0])
		}
	}

	fmt.Fprint(out, "}()")
}

var advance = true

func (pe *PrefixedExpr) GenCode(out io.Writer) {
	fmt.Fprintln(out, "func() (interface{}, error) {")

	if advance && (pe.PrefixOp == AND || pe.PrefixOp == NOT) {
		advance = false
		defer func() {
			advance = true
		}()
	}

	if pe.SuffixedExpr.SuffixOp != 0 {
		fmt.Fprint(out, "// PrefixedExpr: ")
		pe.Print(out)
		fmt.Fprintln(out, "")

		fmt.Fprintln(out, "__peg_pe := func() (interface{}, error) {")
		pe.SuffixedExpr.PrimaryExpr.GenCode(out)
		fmt.Fprintln(out,
			"	return nil, pegErr\n"+
				"}",
		)

		switch pe.SuffixedExpr.SuffixOp {
		case QUESTION: // 0-1
			fmt.Fprintf(out, "return __p.zeroOrOne(__peg_pe)\n")
		case PLUS: // 1-
			fmt.Fprintf(out, "return __p.oneOrMore(__peg_pe)\n")
		case STAR: // 0-
			fmt.Fprintf(out, "return __p.zeroOrMore(__peg_pe)\n")
		}
	} else {
		pe.SuffixedExpr.PrimaryExpr.GenCode(out)
	}

	fmt.Fprint(out, "return nil, pegErr\n}()")
}

func (pe *PrimaryExpr) GenCode(out io.Writer) {
	fmt.Fprint(out, "// PrimaryExpr: ")
	pe.Print(out)
	fmt.Fprintln(out, "")

	switch pe.PrimaryExpr.(type) {
	case *Matcher:
		pe.PrimaryExpr.(*Matcher).GenCode(out)
	case string:
		fmt.Fprintf(out,
			"if _rule_ret, err := __p.rule_%s(); err == nil {\n"+
				"	return _rule_ret, nil\n"+
				"} else {\n"+
				"	return nil, err\n"+
				"}\n",
			pe.PrimaryExpr.(string))
	case *ChoiceExpr:
		pe.PrimaryExpr.(*ChoiceExpr).GenCode(out)
	default:
		panic("type of PrimaryExpr should be *Matcher, string, *ChoiceExpr")
	}
}

func serializeCharRange(chars []*Char) string {
	buf := &bytes.Buffer{}
	for _, c := range chars {
		buf.WriteString(strconv.QuoteRune(c.Start))
		buf.WriteString(", ")
		buf.WriteString(strconv.QuoteRune(c.End))
		buf.WriteString(", ")
	}
	return buf.String()
}

func (m *Matcher) GenCode(out io.Writer) {
	switch m.Matcher.(type) {
	case int:
		fmt.Fprintln(out, "if c := __p.expectDot(); c != \"\" {\n")
		if advance {
			fmt.Fprint(out, "	__p.advance(1)\n")
		}
		fmt.Fprint(out,
			"	return c, nil\n"+
				"}\n",
		)
	case string:
		str := m.Matcher.(string)
		l := utf8.RuneCountInString(str)
		fmt.Fprintf(out, "if __p.expectString(%q, %d) {\n", str, l)
		if advance {
			fmt.Fprintf(out, "	__p.advance(%d)\n", l)
		}
		fmt.Fprintf(out,
			"	return %q, nil\n"+
				"}\n",
			str)
	case *CharRange:
		funcName := "expectChar"
		if m.Matcher.(*CharRange).Not {
			funcName = "expectCharNot"
		}
		fmt.Fprintf(out,
			"if c := __p.%s(%s); c != \"\" {\n",
			funcName,
			serializeCharRange(m.Matcher.(*CharRange).Chars),
		)
		if advance {
			fmt.Fprint(out, "	__p.advance(1)\n")
		}
		fmt.Fprint(out,
			"	return c, nil\n"+
				"}\n",
		)
	default:
		panic("type of Matcher should be int, string, *CharRange")
	}
}
