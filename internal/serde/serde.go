package serde

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
)

func ToRaw(value any) (*runtime.RawExtension, error) {
	valueAsBytes, err := ToBytes(value)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{Raw: valueAsBytes}, nil
}

func ToMap(value any) (map[string]any, error) {
	valueAsBytes, err := ToBytes(value)
	if err != nil {
		return nil, err
	}
	m, err := FromBytes(valueAsBytes, &map[string]any{})
	if err != nil {
		return nil, err
	}
	return *m, nil
}

func ToArray(value any) ([]map[string]any, error) {
	valueAsBytes, err := ToBytes(value)
	if err != nil {
		return nil, err
	}
	a, err := FromBytes(valueAsBytes, &[]map[string]any{})
	if err != nil {
		return nil, err
	}
	return *a, nil
}

func ToBytes(value any) ([]byte, error) {
	valueAsBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return valueAsBytes, nil
}

func FromBytes[T any](source []byte, from *T) (*T, error) {
	err := json.Unmarshal(source, from)
	return from, err
}

func FromMap[T any](source map[string]any, from *T) (*T, error) {
	mapAsBytes, err := ToBytes(source)
	if err != nil {
		return nil, err
	}
	return FromBytes(mapAsBytes, from)
}
