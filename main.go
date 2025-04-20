package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Ciobi0212/gator.git/internal/commands"
	"github.com/Ciobi0212/gator.git/internal/state"
	_ "github.com/lib/pq"
)

func main() {
	state, err := state.GetInitState()

	commands.InitMapCommand()

	if err != nil {
		fmt.Println(fmt.Errorf("error initiating state: %w", err))
		return
	}

	args := os.Args

	if len(args) < 2 {
		// Show help information instead of error when no arguments are provided
		helpCommand := commands.Command{
			Name:   "help",
			Params: []string{},
		}
		err = helpCommand.Run(state)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	commandName := args[1]

	command := commands.Command{
		Name:   commandName,
		Params: args[2:],
	}

	err = command.Run(state)

	var userErr *commands.UserFacingError

	if err != nil {
		if errors.As(err, &userErr) {
			fmt.Println(userErr)
		} else {
			fmt.Println("Internal error, something went wrong")
		}
		os.Exit(1)
	}
}
