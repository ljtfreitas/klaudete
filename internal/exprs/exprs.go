package exprs

import (
	"github.com/nubank/klaudete/internal/exprs/expr"
)

const (
	StartToken = "${"
	EndToken   = "}"
)

type Expression interface {
	Source() string
	Evaluate(map[string]any, ...any) (any, error)
	Dependencies() []string
}

func Parse(expression any, opts ...expr.ExprOption) (Expression, error) {
	expressionAsString, ok := expression.(string)
	if !ok {
		return ScalarExpression{value: expression}, nil
	}

	expressions := expr.SearchExpressions(expressionAsString, opts...)

	if len(expressions) == 0 {
		return ScalarExpression{value: expression}, nil
	}

	if len(expressions) == 1 && isSingleExpression(expressionAsString) {
		return expr.NewExprExpression(expressionAsString)
	}

	return newCompositeExpression(expressionAsString, expressions)

}

func isSingleExpression(expression string) bool {
	return expr.IsSingleExpression(expression)
}

func noDependencies() []string {
	return make([]string, 0)
}

func noArgs() map[string]any {
	return make(map[string]any)
}
