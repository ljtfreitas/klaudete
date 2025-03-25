package generators

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_Generators(t *testing.T) {
	t.Run("We should be able to resolve a list generator", func(t *testing.T) {
		listGeneratorSpecAsBytes, err := json.Marshal(map[string]any{
			"name":   "parameter",
			"values": []string{"one", "two"},
		})
		assert.NoError(t, err)

		generatorList, err := NewGeneratorList(context.TODO(), "list", GeneratorSpec(&runtime.RawExtension{Raw: listGeneratorSpecAsBytes}))
		assert.NoError(t, err)

		assert.Equal(t, "parameter", generatorList.Name)

		expectedVariables := Variables{
			Variable("one"),
			Variable("two"),
		}
		assert.Equal(t, expectedVariables, generatorList.Variables)
	})

	t.Run("We should be able to resolve a data generator", func(t *testing.T) {
		dataGeneratorSpecAsBytes, err := json.Marshal(map[string]any{
			"name": "parameters",
			"values": []map[string]any{
				{
					"name": "one",
				},
				{
					"name": "two",
				},
			},
		})
		assert.NoError(t, err)

		generatorList, err := NewGeneratorList(context.TODO(), "list", GeneratorSpec(&runtime.RawExtension{Raw: dataGeneratorSpecAsBytes}))
		assert.NoError(t, err)

		assert.Equal(t, "parameters", generatorList.Name)

		expectedVariables := Variables{
			Variable(map[string]any{
				"name": "one",
			}),
			Variable(map[string]any{
				"name": "two",
			}),
		}
		assert.Equal(t, expectedVariables, generatorList.Variables)
	})
}
