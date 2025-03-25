package generators

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_ListGenerator(t *testing.T) {
	t.Run("List generator should work as expected", func(t *testing.T) {
		listGenerator := newListGenerator()

		rawSpecAsBytes, err := json.Marshal(map[string]any{
			"name":   "parameter",
			"values": []string{"one", "two"},
		})
		assert.NoError(t, err)

		rawSpec := &runtime.RawExtension{Raw: rawSpecAsBytes}

		name, variables, err := listGenerator.Resolve(nil, GeneratorSpec(rawSpec))
		assert.NoError(t, err)

		assert.Equal(t, "parameter", name)

		expectedVariables := Variables{
			Variable("one"),
			Variable("two"),
		}
		assert.Equal(t, expectedVariables, variables)
	})
}
