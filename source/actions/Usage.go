package actions

import "fmt"
import "os"

func Usage() {

	fmt.Fprint(os.Stdout, "\n")
	fmt.Fprint(os.Stdout, "Usage: dvah <path/to/agent.yml> [path/to/config.yml]\n")
	fmt.Fprint(os.Stdout, "\n")
	fmt.Fprint(os.Stdout, "Arguments:\n")
	fmt.Fprint(os.Stdout, "\n")
	fmt.Fprint(os.Stdout, "  <agent.yml>  string    Path to the agent YAML definition\n")
	fmt.Fprint(os.Stdout, "  [config.yml] string    Optional runtime config (providers, tokens,\n")
	fmt.Fprint(os.Stdout, "                         endpoint URL, initial user prompt)\n")
	fmt.Fprint(os.Stdout, "\n")
	fmt.Fprint(os.Stdout, "The agent runs inside the current working directory (sandbox) and always\n")
	fmt.Fprint(os.Stdout, "starts the terminal UI. There is only one agent. Providers and tokens are\n")
	fmt.Fprint(os.Stdout, "read from ~/.config/dvah/config.yaml and overridden by config.yml.\n")
	fmt.Fprint(os.Stdout, "\n")
	fmt.Fprint(os.Stdout, "Examples:\n")
	fmt.Fprint(os.Stdout, "\n")
	fmt.Fprint(os.Stdout, "  dvah templates/agent.yaml\n")
	fmt.Fprint(os.Stdout, "  dvah templates/agent.yaml templates/config.yaml\n")
	fmt.Fprint(os.Stdout, "  dvah ../agent-outside-sandbox.yml ../config-outside-sandbox.yml\n")
	fmt.Fprint(os.Stdout, "\n")

	os.Exit(1)

}
