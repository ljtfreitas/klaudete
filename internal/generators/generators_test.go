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

		rawSpec := map[string]*runtime.RawExtension{
			"list": &runtime.RawExtension{Raw: listGeneratorSpecAsBytes},
		}

		NewGenerators(context.TODO(), rawSpec)
	})
}
