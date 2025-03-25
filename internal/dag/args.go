package dag

import (
	"fmt"
)

type Args struct {
	all map[string]any
}

func (a *Args) Get(name string) (any, bool) {
	value, exists := a.all[name]
	return value, exists
}

func (a *Args) ToMap() map[string]any {
	return a.all
}

type Arg func(*Args) (*Args, error)

func GeneratorArg(name string, variable any) Arg {
	return func(a *Args) (*Args, error) {
		generator, found := a.all["generator"].(map[string]any)
		if !found {
			generator = make(map[string]any)
		}
		generator[name] = variable
		a.all["generator"] = generator
		return a, nil
	}
}

func EmptyProvisionerArg(name string) Arg {
	return ProvisionerArg(name, map[string]any{
		"status": map[string]any{
			"outputs": make(map[string]any),
		},
	})
}

func ProvisionerArg(name string, provisioner map[string]any) Arg {
	return func(a *Args) (*Args, error) {
		status, found := provisioner["status"].(map[string]any)
		if !found {
			status = make(map[string]any)
		}
		_, found = status["outputs"]
		if !found {
			status["outputs"] = make(map[string]any)
		}
		provisioner["status"] = status

		a.all["provisioner"] = map[string]any{
			name: provisioner,
		}
		return a, nil
	}
}

func NewArg(name string, value any) Arg {
	return func(a *Args) (*Args, error) {
		a.all[name] = value
		return a, nil
	}
}

func NewArgs(args ...Arg) (*Args, error) {
	source := &Args{all: make(map[string]any)}

	for _, arg := range args {
		s, err := arg(source)
		if err != nil {
			return nil, fmt.Errorf("impossible to create arguments: %w", err)
		}
		source = s
	}

	return source, nil
}

func (r *Args) WithArgs(args ...Arg) (*Args, error) {
	source := r

	for _, arg := range args {
		s, err := arg(source)
		if err != nil {
			return nil, fmt.Errorf("impossible to create arguments: %w", err)
		}
		source = s
	}

	return source, nil
}
