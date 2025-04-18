package main

import (
	"fmt"
	"os"

	"github.com/Ciobi0212/gator.git/internal/commands"
	"github.com/Ciobi0212/gator.git/internal/state"
)

func main() {
	state, err := state.GetInitState()

	if err != nil {
		fmt.Println(fmt.Errorf("error initiating state: %w", err))
		return
	}

	args := os.Args

	if len(args) < 2 {
		fmt.Println("Not enough args we're given")
		os.Exit(1)
	}

	commandName := args[1]

	command := commands.Command{
		Name:   commandName,
		Params: args[2:],
	}

	err = command.Run(state)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
