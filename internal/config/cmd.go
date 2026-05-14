package config

type Command struct {
	Name  string
	Usage string
}

var (
	CommandRun = Command{
		Name:  "run",
		Usage: "Run Edge Trust",
	}
	CommandHealthcheck = Command{
		Name:  "healthcheck",
		Usage: "Check daemon heartbeat status",
	}
)

func ParseCommand(raw string) *Command {
	if raw == "" {
		return nil
	}
	commands := commands()
	for i := range commands {
		if commands[i].Name == raw {
			return &commands[i]
		}
	}
	return nil
}

func commands() []Command {
	return []Command{
		CommandHealthcheck,
		CommandRun,
	}
}
