package inventory

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/nubank/klaudete/internal/serde"
	clientv1alpha1 "github.com/nubank/nu-infra-inventory/sdk/pkg/client"
	"github.com/nubank/nu-infra-inventory/sdk/pkg/kcllang"
)

func RunKcl(ctx context.Context, client clientv1alpha1.Client, kcl string) (map[string]any, error) {
	input := strings.NewReader(kcl)
	output := bytes.Buffer{}

	err := kcllang.Run(kcllang.RunArgs{
		Client:       client,
		Input:        input,
		Output:       &output,
		OutputFormat: kcllang.OutputFormatJSON,
		Options:      kcllang.Options{},
	})
	if err != nil {
		return nil, err
	}

	outputAsMap, err := serde.FromJsonString(output.String(), &map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("unable to deserialize kcl output to map: %w", err)
	}

	return *outputAsMap, nil
}
