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

func ProvisionerArg(provisioner map[string]any) Arg {
	return func(a *Args) (*Args, error) {
		if resources, ok := provisioner["resources"].([]map[string]any); ok {
			newResources := make(map[string]any)
			for _, r := range resources {
				newResources[r["name"].(string)] = r
			}
			provisioner["resources"] = newResources
		}
		a.all["provisioner"] = provisioner
		return a, nil
	}
}

func ProvisionerObjectArg(name string, obj map[string]any) Arg {
	return func(a *Args) (*Args, error) {
		if provisioner, ok := a.all["provisioner"].(map[string]any); ok {
			resources, ok := provisioner["resources"].(map[string]any)
			if ok {
				resources[name] = obj
				return a, nil
			}
			provisioner["resources"] = map[string]any{
				name: obj,
			}
			return a, nil
		}
		a.all["provisioner"] = map[string]any{
			"resources": map[string]any{
				name: obj,
			},
		}
		return a, nil
	}
}

func ResourceArg(resource map[string]any) Arg {
	return func(a *Args) (*Args, error) {
		a.all["resource"] = resource
		return a, nil
	}
}

func ResourcesArg(name string, resource map[string]any) Arg {
	return func(a *Args) (*Args, error) {
		resources, ok := a.all["resources"].(map[string]any)
		if !ok {
			resources = make(map[string]any)
		}
		resources[name] = resource
		a.all["resources"] = resources
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
