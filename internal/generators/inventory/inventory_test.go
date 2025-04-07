package inventory

import (
	"context"
	"testing"

	clientv1alpha1 "github.com/nubank/nu-infra-inventory/sdk/pkg/client"
	"github.com/stretchr/testify/assert"
)

func Test_RunKclUsingInventory(t *testing.T) {
	kcl := `import kcl_plugin.inventory
environments = inventory.list_resources("environment")
`
	ctx := context.TODO()

	client, err := clientv1alpha1.New(ctx)
	assert.NoError(t, err)

	output, err := RunKcl(ctx, client, kcl)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)
}
