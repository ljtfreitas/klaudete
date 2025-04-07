package v1alpha1

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceGenerator map[string]*runtime.RawExtension

func (generator ResourceGenerator) Spec() (string, *runtime.RawExtension, error) {
	if generator == nil || len(generator) == 0 {
		return "", nil, nil
	}
	if len(generator) > 1 {
		return "", nil, errors.New("just one generator is allowed.")
	}
	for generatorType, generatorSpec := range generator {
		return generatorType, generatorSpec, nil
	}
	return "", nil, nil
}
