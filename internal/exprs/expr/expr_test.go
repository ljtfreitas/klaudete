package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type ObjectArg struct {
	Object string
}

func Test_SearchExprExpressions(t *testing.T) {
	t.Run("We should be able to detect expressions in a string", func(t *testing.T) {
		matches := SearchExpressions("${i.am.an.expression}")
		assert.Len(t, matches, 1)
		assert.Contains(t, matches, "i.am.an.expression")

		t.Run("...also when we have more than one expression in a single string", func(t *testing.T) {
			matches := SearchExpressions("${i.am.an.expression} ${i.am.another.expression}")
			assert.Len(t, matches, 2)
			assert.Contains(t, matches, "i.am.an.expression")
			assert.Contains(t, matches, "i.am.another.expression")
		})

		t.Run("...and we should be able to exclude expressions starting with variables that we don't want to evaluate", func(t *testing.T) {
			matches := SearchExpressions("${notEvaluatedObject.property}", Exclude("notEvaluatedObject"))
			assert.Empty(t, matches, 0)

			t.Run("...even when we have composed expressions", func(t *testing.T) {
				matches := SearchExpressions("${notEvaluatedObject.property} ${evaluatedObject.property}", Exclude("notEvaluatedObject"))
				assert.Len(t, matches, 1)
				assert.Contains(t, matches, "evaluatedObject.property")
			})
		})
	})
}

func Test_IsSingleExpression(t *testing.T) {
	t.Run("We should be able to detect when a candidate is a single expressions", func(t *testing.T) {
		t.Run("...for the simplest cases, when we have just one expression", func(t *testing.T) {
			assert.True(t, IsSingleExpression("${single.expression}"))
		})

		t.Run("...and for more complex cases, like composed expressions. these examples should not be considered single expressions", func(t *testing.T) {
			assert.False(t, IsSingleExpression("ola, ${single.expression}"))
			assert.False(t, IsSingleExpression("ola, ${single.expression}, hello"))

			t.Run("...a candidate with two expr expressions also is not a single expression", func(t *testing.T) {
				assert.False(t, IsSingleExpression("${single.expression} ${another.expression}"))
				assert.False(t, IsSingleExpression("ola ${single.expression}, hello ${another.expression}"))
			})
		})
	})
}

func Test_ExprExpression(t *testing.T) {

	t.Run("We should be able to eval a constant expression", func(t *testing.T) {
		expression, err := NewExprExpression(`${"sample"}`)

		assert.NoError(t, err)
		assert.Equal(t, `"sample"`, expression.Source())

		r, err := expression.Evaluate(make(map[string]any))

		assert.NoError(t, err)
		assert.Equal(t, "sample", r)
	})

	t.Run("We should be able to eval an array expression", func(t *testing.T) {
		expression, err := NewExprExpression(`${sample[1]}`)

		assert.NoError(t, err)
		assert.Equal(t, "sample[1]", expression.Source())

		variables := map[string]any{
			"sample": []string{
				"hello",
				"world",
			},
		}

		r, err := expression.Evaluate(variables)

		assert.NoError(t, err)
		assert.Equal(t, "world", r)
	})

	t.Run("We should be able to eval an object expression", func(t *testing.T) {

		t.Run("We can use a map as variable", func(t *testing.T) {
			expression, err := NewExprExpression("${i.am.an.object}")

			assert.NoError(t, err)
			assert.Equal(t, "i.am.an.object", expression.Source())

			variables := map[string]any{
				"i": map[string]any{
					"am": map[string]any{
						"an": map[string]any{
							"object": "i am an object!",
						},
					},
				},
			}

			r, err := expression.Evaluate(variables)

			assert.NoError(t, err)
			assert.Equal(t, "i am an object!", r)
		})

		t.Run("We can use a struct as variable", func(t *testing.T) {
			expression, err := NewExprExpression("${i.am.an.Object}")

			assert.NoError(t, err)
			assert.Equal(t, "i.am.an.Object", expression.Source())

			variables := map[string]any{
				"i": map[string]any{
					"am": map[string]any{
						"an": ObjectArg{
							Object: "i am an object!",
						},
					},
				},
			}

			r, err := expression.Evaluate(variables)

			assert.NoError(t, err)
			assert.Equal(t, "i am an object!", r)
		})

		t.Run("We can use a map whose keys contains kebab-names", func(t *testing.T) {
			expression, err := NewExprExpression("${i['am-an'].object}")

			assert.NoError(t, err)
			assert.Equal(t, "i['am-an'].object", expression.Source())

			variables := map[string]any{
				"i": map[string]any{
					"am-an": map[string]any{
						"object": "i am an object!",
					},
				},
			}

			r, err := expression.Evaluate(variables)

			assert.NoError(t, err)
			assert.Equal(t, "i am an object!", r)
		})
	})
}
