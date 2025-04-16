package exprs

import "fmt"

type ScalarExpression struct {
	value any
}

func (e ScalarExpression) Source() string {
	return fmt.Sprint(e.value)
}

func (e ScalarExpression) Evaluate(map[string]any, ...any) (any, error) {
	return e.value, nil
}

func (e ScalarExpression) Dependencies() []string {
	return noDependencies()
}
