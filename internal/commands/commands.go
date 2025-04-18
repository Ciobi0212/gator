package commands

import (
	"errors"
	"fmt"

	"github.com/Ciobi0212/gator.git/internal/state"
)

type Command struct {
	Name   string
	Params []string
}

var mapCommands = map[string]func(*state.AppState, Command) error{
	"login": handleLogin,
}

func handleLogin(state *state.AppState, cmd Command) error {
	if len(cmd.Params) != 1 {
		return errors.New("login command expects 1 param : <username>")
	}

	username := cmd.Params[0]

	err := state.Cfg.SetUser(username)

	if err != nil {
		return fmt.Errorf("err setting user: %w", err)
	}

	fmt.Printf("Current user is %s\n", username)

	return nil
}

func (c *Command) Run(state *state.AppState) error {
	callback, ok := mapCommands[c.Name]

	if !ok {
		return errors.New("uknown command")
	}

	err := callback(state, *c)

	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	return nil
}
