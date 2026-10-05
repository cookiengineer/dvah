package tty

import "dvah/adapters"
import "dvah/engine"
import "dvah/schemas"
import "dvah/tools"
import "dvah/types"
import "bufio"
import "fmt"
import "os"
import "os/signal"
import "strings"
import "syscall"
import "time"

type Client struct {
	Renderer *Renderer
	Session  *engine.Session
	Role     string
}

func NewClient(agent *types.Agent, config *types.Config) *Client {

	var session *engine.Session = nil

	recovery := engine.NewRecovery(config.Sandbox)

	if recovery.HasBackup() {

		session = recovery.RestoreSession()

		if session != nil {
			session.Console.Info("Restored Session from Backup")
		} else {
			session = engine.NewSession(agent, config)
			session.Console.Warn("Could not restore Session from Backup")
		}

	} else {
		session = engine.NewSession(agent, config)
	}

	renderer := NewRenderer(session)

	if agent.HasTools() {

		toolset := tools.Toolset(
			config.Sandbox,
			agent.AllowedPrograms,
			agent.AllowedTools,
			session.Permissions,
		)

		if len(toolset) > 0 {

			for _, tool := range toolset {
				session.SetTool(tool)
			}

		}

	}

	if config.HasProvider(agent.Model) {

		adapterset := adapters.Adapterset(
			config.ResolveURL(agent.Model, ""),
			config.ResolveModel(agent.Model),
		)

		if len(adapterset) > 0 {

			for _, adapter := range adapterset {
				session.SetAdapter(adapter)
			}

		}

	}

	if config.GetPrompt() != "" {
		session.Console.Info("Initial prompt attached to Session")
	}

	return &Client{
		Renderer: renderer,
		Session:  session,
		Role:     "user",
	}

}

func (client *Client) Destroy() {

	if client.Session != nil {
		client.Session.Recovery.BackupSession(client.Session)
	}

	if client.Renderer != nil {
		client.Renderer.Destroy()
	}

}

func (client *Client) Init() {

	signals := make(chan os.Signal, 1)

	signal.Notify(
		signals,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	if client.Session != nil {

		go func() {

			err := client.Session.Init()

			if err != nil {
				fmt.Fprintf(os.Stderr, "\nError: %s\n", err.Error())
			}

		}()

	}

	go func() {
		client.InputLoop()
		signals<-syscall.SIGINT
	}()

	if client.Renderer != nil {
		go client.Renderer.RenderLoop()
	}

	select {
	case sig := <-signals:

		switch sig {
		case syscall.SIGINT:

			client.Destroy()
			fmt.Fprintf(os.Stdout, "Received signal: %s\n", "SIGINT")

			time.Sleep(1 * time.Second)
			os.Exit(0)

		case syscall.SIGTERM:

			client.Destroy()
			fmt.Fprintf(os.Stdout, "Received signal: %s\n", "SIGTERM")

			time.Sleep(1 * time.Second)
			os.Exit(0)

		default:

			client.Destroy()
			fmt.Printf("Received signal: %s\n", sig.String())

			time.Sleep(1 * time.Second)
			os.Exit(0)

		}

	}

}

func (client *Client) InputLoop() {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		role   := client.Role
		prompt := strings.TrimSpace(scanner.Text())

		if prompt != "" && client.Session != nil {

			if role == "user" || role == "assistant" {

				if strings.HasPrefix(prompt, "/") && strings.Contains(prompt, " ") && !strings.Contains(prompt, "\n") {

					command := types.ParseCommand(prompt)

					if command != nil {
						client.Session.CallTool("", command.Name, command.Method, command.Arguments)
					}

				} else {

					go func() {

						err := client.Session.SendChatRequest(schemas.Message{
							Role:    role,
							Content: prompt,
						})

						if err != nil {

							fmt.Fprintf(os.Stderr, "\nFatal Error: %s\n", err.Error())
							os.Exit(1)

						}

					}()

				}

			}

		}

	}

}

func (client *Client) SetRole(role string) {

	if role == "user" || role == "assistant" {
		client.Role = role
		client.Renderer.Role = role
	}

}
