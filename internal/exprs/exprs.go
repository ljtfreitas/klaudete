package exprs

import (
	"strings"

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

func Parse(expression any) (Expression, error) {
	expressionAsString, ok := expression.(string)
	if !ok {
		return ScalarExpression{value: expression}, nil
	}

	expressions := expr.SearchExpressions(expressionAsString)

	if len(expressions) == 0 {
		return ScalarExpression{value: expression}, nil
	}

	if len(expressions) == 1 && strings.HasPrefix(expressionAsString, StartToken) && strings.HasSuffix(expressionAsString, EndToken) {
		return expr.NewExprExpression(expressionAsString)
	}

	return newCompositeExpression(expressionAsString, expressions)

}

func noDependencies() []string {
	return make([]string, 0)
}

func noArgs() map[string]any {
	return make(map[string]any)
}
