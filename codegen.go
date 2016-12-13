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
func main() {
	src, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, "Input:----\n"+ string(src)+ "\n----------")
	fmt.Println(Parse([]rune(string(src))))
}

func Parse(src []rune) interface{} {
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

func (__p *parser) expectDot() rune {
	return __p.src[__p.n]
}

func (__p *parser) expectString(str string, l int) bool {
	if __p.n + l < len(__p.src) && str == string(__p.src[__p.n:__p.n+l]) {
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
`

func (tree *Tree) GenCode(out io.Writer) {
	fmt.Fprintf(out, "package %s\n", "main")

	fmt.Fprintf(out, header,
		tree.RuleList[0].Name)

	fmt.Fprint(out, tree.Grammar.Code)

	for _, r := range tree.RuleList {
		r.GenCode(out)
	}
}

func (r *Rule) GenCode(out io.Writer) {
	fmt.Fprint(out, "// Rule: ")
	r.Print(out)
	fmt.Fprintln(out, "")

	fmt.Fprintf(out, "func (__p *parser) rule_%s() interface{} {\n", r.Name)

	r.ChoiceExpr.GenCode(out)

	fmt.Fprintln(out, "return nil")

	fmt.Fprintln(out, "}\n")
}

func (ce *ChoiceExpr) GenCode(out io.Writer) {
	fmt.Fprintln(out, "var __peg_n int")
	for _, ae := range ce.ActionExprs {
		fmt.Fprintln(out, "__peg_n = __p.n")
		fmt.Fprintf(out, "if __ae_ret := ")
		ae.GenCode(out)
		fmt.Fprintf(out,
			"; __ae_ret != nil {\n"+
				"	return __ae_ret\n"+
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

func (ae *ActionExpr) GenCode(out io.Writer) {
	fmt.Fprint(out, "func() interface{} {\n")

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

		fmt.Fprint(out, "if __pe_ret := ")

		le.PrefixedExpr.GenCode(out)

		valueVarOrEmpty := "__pe_ret"
		not := "!"
		if le.PrefixedExpr.PrefixOp == AND || le.PrefixedExpr.PrefixOp == NOT {
			valueVarOrEmpty = "nil"
		}
		if le.PrefixedExpr.PrefixOp == NOT {
			not = "="
		}
		fmt.Fprintf(out,
			"; __pe_ret %s= nil {\n"+
				"	%s = %s\n"+
				"} else {\n"+
				"	return nil\n"+
				"}\n",
			not,
			varName,
			valueVarOrEmpty,
		)
	}
	if ae.Code != "" {
		fmt.Fprintf(out, "return func() interface{} {\n%s\n}()\n", ae.Code)
	} else {
		if len(vars) > 1 {
			fmt.Fprintf(out, "return [...]interface{}{%s}\n", strings.Join(vars, ", "))
		} else {
			fmt.Fprintf(out, "return %s\n", vars[0])
		}
	}

	fmt.Fprint(out, "}()")
}

var advance = true

func (pe *PrefixedExpr) GenCode(out io.Writer) {
	fmt.Fprintln(out, "func() interface{} {")

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

		fmt.Fprintln(out, "__peg_pe := func() interface{} {")
		pe.SuffixedExpr.PrimaryExpr.GenCode(out)
		fmt.Fprintln(out,
			"	return nil\n"+
				"}",
		)

		switch pe.SuffixedExpr.SuffixOp {
		case QUESTION: // 0-1
			fmt.Fprintf(out,
				"var __peg_ret interface{} = \"\"\n"+
					"if _r := __peg_pe(); _r != nil {\n"+
					"	__peg_ret = _r\n"+
					"}\n"+
					"return __peg_ret\n",
			)
		case PLUS: // 1-
			fmt.Fprintf(out,
				"var __peg_ret []interface{}\n"+
					"if _r := __peg_pe(); _r != nil {\n"+
					"	__peg_ret = []interface{}{_r}\n"+
					"} else {\n"+
					"	return nil\n"+
					"}\n"+
					"for {\n"+
					"	if _r := __peg_pe(); _r != nil {\n"+
					"		__peg_ret = append(__peg_ret, _r)\n"+
					"	} else {\n"+
					"		break\n"+
					"	}\n"+
					"}\n"+
					"if len(__peg_ret) > 0 {\n"+
					"	return __peg_ret\n"+
					"} else {\n"+
					"	return nil\n"+
					"}\n",
			)
		case STAR: // 0-
			fmt.Fprintf(out,
				"__peg_ret := []interface{}{}\n"+
					"for {\n"+
					"	if _r := __peg_pe(); _r != nil {\n"+
					"		__peg_ret = append(__peg_ret, _r)\n"+
					"	} else {\n"+
					"		break\n"+
					"	}\n"+
					"}\n"+
					"if len(__peg_ret) >= 0 {\n"+
					"	return __peg_ret\n"+
					"} else {\n"+
					"	return nil\n"+
					"}\n",
			)
		}
	} else {
		pe.SuffixedExpr.PrimaryExpr.GenCode(out)
	}

	fmt.Fprint(out, "return nil\n}()")
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
			"if _rule_ret := __p.rule_%s(); _rule_ret != nil {\n"+
				"	return _rule_ret\n"+
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
		fmt.Fprintln(out, "return __p.expectDot()")
	case string:
		str := m.Matcher.(string)
		l := utf8.RuneCountInString(str)
		fmt.Fprintf(out, "if __p.expectString(%q, %d) {\n", str, l)
		if advance {
			fmt.Fprintf(out, "	__p.advance(%d)\n", l)
		}
		fmt.Fprintf(out,
			"	return %q\n"+
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
			"	return c\n"+
				"}\n",
		)
	default:
		panic("type of Matcher should be int, string, *CharRange")
	}
}
