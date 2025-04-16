package expr

import (
	"fmt"
	"regexp"

	"github.com/expr-lang/expr"
)

var (
	exprExpressionRe      = regexp.MustCompile(`\$\{([^}]+)\}`)
	resourcesExpressionRe = regexp.MustCompile(`(resources\.[^.]+)\.`)
	generatorExpressionRe = regexp.MustCompile(`(generator\.[^.]+)\.`)

	resourcesEscapedExpressionRe = regexp.MustCompile(`(resources)\["([^"]+)"\]`)
	generatorEscapedExpressionRe = regexp.MustCompile(`(generator)\["([^"]+)"\]`)

	checkSingleExprRe = regexp.MustCompile(`^\$\{[^}]+\}$`)
)

type ExprOption func(string) string

func Exclude(variableToExclude string) ExprOption {
	return func(expr string) string {
		pattern := fmt.Sprintf(`^%s`, regexp.QuoteMeta(variableToExclude))
		matched, err := regexp.MatchString(pattern, expr)
		if err != nil {
			return expr
		}
		if matched {
			return ""
		}
		return expr
	}
}

func Only(variableToExpand string) ExprOption {
	return func(expr string) string {
		pattern := fmt.Sprintf(`^%s`, regexp.QuoteMeta(variableToExpand))
		matched, err := regexp.MatchString(pattern, expr)
		if err != nil {
			return expr
		}
		if matched {
			return expr
		}
		return ""
	}
}

func IsSingleExpression(expr string) bool {
	return checkSingleExprRe.MatchString(expr)
}

func SearchExpressions(expression string, opts ...ExprOption) []string {
	matches := exprExpressionRe.FindAllStringSubmatch(expression, -1)

	applyOptions := func(source string) string {
		for _, o := range opts {
			source = o(source)
		}
		return source
	}

	expressions := make([]string, 0)
	for _, m := range matches {
		expression := applyOptions(m[1])
		if expression != "" {
			expressions = append(expressions, expression)
		}
	}

	return expressions
}

type ExprExpression string

func NewExprExpression(source string) (ExprExpression, error) {
	matches := exprExpressionRe.FindStringSubmatch(source)

	if len(matches) == 0 {
		return ExprExpression(""), fmt.Errorf("invalid Expr expression: %s", source)
	}

	expression := matches[1]

	return ExprExpression(expression), nil
}

func (e ExprExpression) Source() string {
	return string(e)
}

func (e ExprExpression) Dependencies() []string {
	dependencies := make([]string, 0)

	matches := resourcesExpressionRe.FindStringSubmatch(e.Source())
	if len(matches) != 0 {
		dependencies = append(dependencies, matches[1])
	}

	matches = generatorExpressionRe.FindStringSubmatch(e.Source())
	if len(matches) != 0 {
		dependencies = append(dependencies, matches[1])
	}

	matches = resourcesEscapedExpressionRe.FindStringSubmatch(e.Source())
	if len(matches) > 2 {
		dependencies = append(dependencies, fmt.Sprintf("%s.%s", matches[1], matches[2]))
	}

	matches = generatorEscapedExpressionRe.FindStringSubmatch(e.Source())
	if len(matches) > 2 {
		dependencies = append(dependencies, fmt.Sprintf("%s.%s", matches[1], matches[2]))
	}

	return dependencies
}

func (e ExprExpression) Evaluate(args map[string]any, options ...any) (any, error) {
	exprOpts := make([]expr.Option, 0)

	if args != nil {
		exprOpts = append(exprOpts, expr.Env(args))
	}

	for _, option := range options {
		if exprOption, ok := option.(expr.Option); ok {
			exprOpts = append(exprOpts, exprOption)
		}
	}

	source := e.Source()

	program, err := expr.Compile(source, exprOpts...)
	if err != nil {
		return "", fmt.Errorf("failure compiling expression %s: %w", source, err)
	}

	value, err := expr.Run(program, args)
	if err != nil {
		return "", fmt.Errorf("failure evaluating expression %s: %w", source, err)
	}

	return value, nil
}

func Function(name string, fn func(...any) (any, error)) expr.Option {
	return expr.Function(name, fn)
}
