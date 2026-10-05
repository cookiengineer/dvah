package tools

import "dvah/types"
import "strings"

func Toolset(sandbox string, allowed_programs []string, allowed_tools []string, permissions *types.PermissionStore) []types.Tool {

	filtered_by_methods := make(map[string][]string, 0)
	filtered_by_tool    := make(map[string]types.Tool)

	for _, tool_name := range allowed_tools {

		if strings.Contains(tool_name, ".") {

			tmp1 := strings.TrimSpace(tool_name[0:strings.Index(tool_name, ".")])
			tmp2 := strings.TrimSpace(tool_name[strings.Index(tool_name, ".")+1:])

			name   := strings.ToLower(tmp1)
			method := strings.ToUpper(tmp2[0:1]) + strings.ToLower(tmp2[1:])

			_, ok := filtered_by_methods[name]

			if ok == false {
				filtered_by_methods[name] = make([]string, 0)
			}

			filtered_by_methods[name] = append(filtered_by_methods[name], method)

		}

	}

	for tool_name, allowed_methods := range filtered_by_methods {

		var tool types.Tool = nil

		switch tool_name {
		case "files":
			tool = NewFiles(allowed_methods, sandbox)
		case "permissions":
			tool = NewPermissions(allowed_methods, permissions)
		case "programs":
			tool = NewPrograms(allowed_methods, sandbox, allowed_programs)
		case "skills":
			tool = NewSkills(allowed_methods, sandbox, allowed_programs, allowed_tools)
		case "websites":
			tool = NewWebsites(allowed_methods, sandbox)
		}

		if tool != nil {
			filtered_by_tool[tool_name] = tool
		}

	}

	result := make([]types.Tool, 0)

	for _, tool := range filtered_by_tool {
		result = append(result, tool)
	}

	return result

}
