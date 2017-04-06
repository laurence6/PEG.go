package peg

import (
	"errors"
	"fmt"
)

var checkers = [...]func(*Tree) []error{
	func(tree *Tree) (errs []error) {
		rs := map[string]struct{}{}

		// Dup rule
		for _, r := range tree.RuleList {
			if _, ok := rs[r.Name]; !ok {
				rs[r.Name] = struct{}{}
			} else {
				errs = append(errs, errors.New(
					fmt.Sprintf("Dup rule %q", r.Name),
				))
			}
		}

		// Rule undefined
		stack := []*ChoiceExpr{}

		checkCE := func(ce *ChoiceExpr) {
			for _, ae := range ce.ActionExprs {
				for _, le := range ae.SeqExpr.LabeledExprs {
					pe := le.PrefixedExpr.SuffixedExpr.PrimaryExpr.PrimaryExpr
					switch pe.(type) {
					case string:
						rn := pe.(string)
						if _, ok := rs[rn]; !ok {
							errs = append(errs, errors.New(
								fmt.Sprintf("Rule %q undefined", rn),
							))
						}
					case *ChoiceExpr:
						stack = append(stack, pe.(*ChoiceExpr))
					}
				}
			}
		}

		for _, r := range tree.RuleList {
			stack = append(stack, r.ChoiceExpr)
		}

		for len(stack) > 0 {
			checkCE(stack[0])
			stack = stack[1:]
		}

		return
	},
}

func Check(tree *Tree) []error {
	errs := []error{}

	for _, checker := range checkers {
		_errs := checker(tree)
		if _errs != nil && len(_errs) > 0 {
			errs = append(errs, _errs...)
		}
	}

	return errs
}
