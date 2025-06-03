package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/curtisbraxdale/blog-gator/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	value, exists := c.commandMap[cmd.name]
	if exists {
		err := value(s, cmd)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Command not found.")
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}

func main() {
	fig, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config file.")
		return
	}

	appState := state{config: &fig}
	cliCommands := commands{make(map[string]func(*state, command) error)}
	cliCommands.register("login", handlerLogin)

	cliArguments := os.Args
	if len(cliArguments) < 2 {
		err = fmt.Errorf("Not enough arguments.")
		fmt.Printf("Error Found: %v\n", err)
		os.Exit(1)
	}

	commandName := cliArguments[1]
	commandArguments := cliArguments[2:]
	newCommand := command{name: commandName, arguments: commandArguments}
	err = cliCommands.run(&appState, newCommand)
	if err != nil {
		fmt.Printf("Error Found: %v\n", err)
		os.Exit(1)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("No Arguments")
	}
	username := cmd.arguments[0]
	err := s.config.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("User has been set to: %v", username)
	return nil
}
