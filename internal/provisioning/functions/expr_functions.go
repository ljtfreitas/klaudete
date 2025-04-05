package functions

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Secrets(ctx context.Context, client client.Client) func(...any) (any, error) {
	return func(args ...any) (any, error) {
		name := args[0].(string)
		namespace := args[1].(string)

		secret := &corev1.Secret{}
		if err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, secret); err != nil {
			return nil, nil
		}
		secretAsJson, err := json.Marshal(secret)
		if err != nil {
			return nil, err
		}
		secretAsMap := make(map[string]any)
		err = json.Unmarshal(secretAsJson, &secretAsMap)
		if err != nil {
			return nil, err
		}
		data := make(map[string]any)
		for name, value := range secret.Data {
			data[name] = string(value)
		}
		secretAsMap["data"] = data
		return secretAsMap, nil
	}
}
