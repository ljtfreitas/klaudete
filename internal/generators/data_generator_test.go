package generators

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_DataGenerator(t *testing.T) {
	t.Run("Data generator should work as expected", func(t *testing.T) {
		dataGenerator := newDataGenerator()

		rawSpecAsBytes, err := json.Marshal(map[string]any{
			"name": "cats",
			"values": []map[string]any{
				{
					"name": "Puka",
				},
				{
					"name": "Fiona",
				},
			},
		})
		assert.NoError(t, err)

		rawSpec := &runtime.RawExtension{Raw: rawSpecAsBytes}

		name, variables, err := dataGenerator.Resolve(nil, GeneratorSpec(rawSpec))
		assert.NoError(t, err)

		assert.Equal(t, "cats", name)

		expected := Variables{
			Variable(map[string]any{
				"name": "Puka",
			}),
			Variable(map[string]any{
				"name": "Fiona",
			}),
		}

		assert.Equal(t, expected, variables)
	})
}
