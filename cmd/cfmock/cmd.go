package main

import "fmt"

type Command struct {
	Name    string
	Usage   string
	Handler func()
}

var (
	CommandSet = Command{
		Name:    "set",
		Usage:   "Set exact IP ranges",
		Handler: handleSet,
	}
	CommandRandom = Command{
		Name:    "random",
		Usage:   "Generate random IP ranges",
		Handler: handleRandom,
	}
	CommandClear = Command{
		Name:    "clear",
		Usage:   "Clear all IP ranges",
		Handler: handleClear,
	}
	CommandDelete = Command{
		Name:    "delete",
		Usage:   "Delete ips.json response file",
		Handler: handleDelete,
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
		CommandSet,
		CommandRandom,
		CommandClear,
		CommandDelete,
	}
}

func handleSet() {
	response := runSet()

	fmt.Println("IPs set successfully")

	response.Result.Print()
}

func handleRandom() {
	response := runRandom()

	fmt.Println("IPs randomized successfully")

	response.Result.Print()
}

func handleClear() {
	response := runClear()

	fmt.Println("IPs cleared successfully")

	response.Result.Print()
}

func handleDelete() {
	runDelete()

	fmt.Println("Response file deleted successfully")
}
