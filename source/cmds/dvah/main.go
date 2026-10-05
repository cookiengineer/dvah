package main

import "dvah/actions"
import "dvah/agents"
import "dvah/providers"
import "dvah/types"
import "fmt"
import "os"

func main() {

	if len(os.Args) < 2 {
		actions.Usage()
		return
	}

	agent, err0 := agents.LoadAgent(os.Args[1])

	if err0 != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err0.Error())
		os.Exit(1)
	}

	cwd, err1 := os.Getwd()

	if err1 != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err1.Error())
		os.Exit(1)
	}

	config := types.NewConfig(
		agent.Name,
		agent.Model,
		"",
		agent.Temperature,
		cwd,
		nil,
		false,
	)

	// Global runtime defaults, then the optional per-invocation config.yml
	config.Merge(types.GlobalConfig)

	if len(os.Args) > 2 {

		file_config, err2 := types.LoadConfig(os.Args[2])

		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err2.Error())
			os.Exit(1)
		}

		config.Merge(file_config)

	}

	// Resolve provider presets (e.g. deepseek aliases) for the agent model
	if preset := providers.NewProvider(agent.Model); preset != nil {

		provider, ok := config.Providers[agent.Model]

		if ok == false {
			provider = *preset
		} else {
			if provider.URL == nil {
				provider.URL = preset.URL
			}
			if provider.Alias == "" {
				provider.Alias = preset.Alias
			}
			if provider.Pricing == (types.Pricing{}) {
				provider.Pricing = preset.Pricing
			}
		}

		config.Providers[agent.Model] = provider

	}

	agent.Sandbox = config.Sandbox

	err3 := os.MkdirAll(config.Sandbox, 0755)

	if err3 != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err3.Error())
		os.Exit(1)
	}

	actions.Terminal(agent, config, "user")

}
