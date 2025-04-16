package inventory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RunKclUsingInventory(t *testing.T) {
	kcl := `import kcl_plugin.inventory
environments = inventory.list_resources("environment")
`
	ctx := context.TODO()

	inventory, err := NewInventoryClient()
	assert.NoError(t, err)

	output, err := inventory.RunKcl(ctx, kcl)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)
}
