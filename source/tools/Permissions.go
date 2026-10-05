package tools

import "dvah/schemas"
import "dvah/types"
import "fmt"
import "slices"
import "strings"

type Permissions struct {
	Methods []string
	store   *types.PermissionStore
}

func NewPermissions(methods []string, store *types.PermissionStore) *Permissions {

	return &Permissions{
		Methods: methods,
		store:   store,
	}

}

func (tool *Permissions) Name() string {
	return "permissions"
}

func (tool *Permissions) Call(method string, arguments map[string]interface{}) (string, error) {

	if tool.HasMethod(method) == true {

		if method == "List" {

			return tool.List()

		} else if method == "Request" {

			name,   ok1 := arguments["tool"].(string)
			action, ok2 := arguments["method"].(string)
			scope,  ok3 := arguments["scope"].(string)

			if ok1 == true && ok2 == true && ok3 == true {
				return tool.Request(name, action, scope)
			} else {
				return "", fmt.Errorf("permissions.%s: Invalid parameters \"tool\", \"method\" and \"scope\" must be strings.", method)
			}

		} else {
			return "", fmt.Errorf("permissions.%s: Invalid method.", method)
		}

	} else {
		return "", fmt.Errorf("permissions.%s: Method not allowed.", method)
	}

}

func (tool *Permissions) GetContent(id string) (any, error) {
	return nil, nil
}

func (tool *Permissions) GetContentIdentifiers() []string {
	return []string{}
}

func (tool *Permissions) HasMethod(method string) bool {
	return slices.Contains(tool.Methods, method) == true
}

func (tool *Permissions) Request(name string, method string, scope string) (string, error) {

	if tool.store == nil {
		return "", fmt.Errorf("permissions.Request: No permission store configured.")
	}

	name = strings.ToLower(strings.TrimSpace(name))

	tool.store.Request(name, method, scope)

	return fmt.Sprintf("permissions.Request: Granted %s.%s for scope \"%s\".", name, method, scope), nil

}

func (tool *Permissions) List() (string, error) {

	if tool.store == nil {
		return "", fmt.Errorf("permissions.List: No permission store configured.")
	}

	grants := tool.store.Grants()

	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("permissions.List: %d grants.", len(grants)))

	for _, grant := range grants {
		lines = append(lines, fmt.Sprintf("- %s.%s (scope: %s)", grant.Tool, grant.Method, grant.Scope))
	}

	return strings.Join(lines, "\n"), nil

}

func (tool *Permissions) Schemas() []schemas.Tool {

	result := make([]schemas.Tool, 0)

	for _, method := range tool.Methods {

		for _, schema := range PermissionsSchema {

			if schema.Function.Name == fmt.Sprintf("%s.%s", tool.Name(), method) {
				result = append(result, schema)
			}

		}

	}

	return result

}
