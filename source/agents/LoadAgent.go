package agents

import "dvah/schemas"
import "dvah/types"
import "fmt"
import "os"
import "path/filepath"
import "strings"

func LoadAgent(path string) (*types.Agent, error) {

	resolved, err0 := filepath.Abs(strings.TrimSpace(path))

	if err0 != nil {
		return nil, fmt.Errorf("LoadAgent: Invalid path \"%s\"", path)
	}

	stat, err1 := os.Stat(resolved)

	if err1 != nil {
		return nil, fmt.Errorf("LoadAgent: Cannot read \"%s\": %s", path, err1.Error())
	} else if stat.IsDir() == true {
		return nil, fmt.Errorf("LoadAgent: \"%s\" is a directory, expected an agent.yml file", path)
	}

	data, err2 := os.ReadFile(resolved)

	if err2 != nil {
		return nil, fmt.Errorf("LoadAgent: %s", err2.Error())
	}

	agent, err3 := types.ParseAgent(data)

	if err3 != nil {
		return nil, fmt.Errorf("LoadAgent: %s", err3.Error())
	}

	if strings.TrimSpace(agent.Name) == "" {
		return nil, fmt.Errorf("LoadAgent: Missing agent \"name\"")
	}

	if strings.TrimSpace(agent.Model) == "" {
		return nil, fmt.Errorf("LoadAgent: Missing agent \"model\"")
	}

	if strings.TrimSpace(agent.Role) == "" {
		agent.Role = "agent"
	}

	agent.Prompt   = render_prompt(agent.Name, agent.Role, agent.Prompt)
	agent.Messages = render_messages(agent.Prompt)
	agent.Status   = "working"
	agent.StartedAt = schemas.NewDatetime()

	return agent, nil

}
