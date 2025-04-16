package patches

import (
	"testing"

	"github.com/nubank/klaudete/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func Test_ResourcePatches(t *testing.T) {
	t.Run("We should be able to apply properties to nested paths", func(t *testing.T) {
		patch := v1alpha1.ResourcePatch{
			From: "ola",
			To:   "status.properties.message",
		}

		statusAsMap := ApplyTo(patch, make(map[string]any))
		assert.Contains(t, statusAsMap, "status")

		status := statusAsMap["status"].(map[string]any)
		assert.Contains(t, status, "properties")

		properties := status["properties"].(map[string]any)
		assert.Contains(t, properties, "message")

		message := properties["message"]
		assert.Equal(t, patch.From, message)
	})

	t.Run("We should be able to apply properties in a single path", func(t *testing.T) {
		patch := v1alpha1.ResourcePatch{
			From: "ola",
			To:   "whatever",
		}

		output := ApplyTo(patch, make(map[string]any))
		assert.Contains(t, output, "whatever")

		value := output["whatever"]
		assert.Equal(t, patch.From, value)
	})

	t.Run("We should be able to keep map state", func(t *testing.T) {
		patch := v1alpha1.ResourcePatch{
			From: "ola",
			To:   "status.properties.ola",
		}

		source := map[string]any{
			"status": map[string]any{
				"properties": map[string]any{
					"hello": "hello",
				},
			},
		}

		statusAsMap := ApplyTo(patch, source)
		assert.Contains(t, statusAsMap, "status")

		status := statusAsMap["status"].(map[string]any)
		assert.Contains(t, status, "properties")

		properties := status["properties"].(map[string]any)
		assert.Contains(t, properties, "ola")
		assert.Contains(t, properties, "hello")

		ola := properties["ola"]
		assert.Equal(t, patch.From, ola)

		hello := properties["hello"]
		assert.Equal(t, "hello", hello)
	})
}
