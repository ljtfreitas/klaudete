package exprs

import (
	"fmt"
	"strings"

	"github.com/nubank/klaudete/internal/exprs/expr"
)

type CompositeExpression struct {
	source      string
	expressions []Expression
}

func newCompositeExpression(expression string, expressions []string) (CompositeExpression, error) {
	checkedExpressions := make([]Expression, 0)
	for _, e := range expressions {
		checkedExpressions = append(checkedExpressions, expr.ExprExpression(e))
	}

	return CompositeExpression{source: expression, expressions: checkedExpressions}, nil
}

func (e CompositeExpression) Source() string {
	return e.source
}

func (e CompositeExpression) Evaluate(args map[string]any, opts ...any) (any, error) {
	s := e.source

	for _, expression := range e.expressions {
		r, err := expression.Evaluate(args, opts...)
		if err != nil {
			return "", err
		}

		fragment := StartToken + expression.Source() + EndToken
		s = strings.Replace(s, fragment, fmt.Sprintf("%s", r), -1)
	}
	return s, nil
}

func (e CompositeExpression) Dependencies() []string {
	dependencies := make([]string, 0, len(e.expressions))
	for _, expression := range e.expressions {
		dependencies = append(dependencies, expression.Dependencies()...)
	}
	return dependencies
}
