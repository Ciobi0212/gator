package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Ciobi0212/gator.git/internal/database"
	"github.com/Ciobi0212/gator.git/internal/state"
	"github.com/google/uuid"
)

type Command struct {
	Name   string
	Params []string
}

var mapCommands = map[string]func(*state.AppState, []string) error{
	"login":    handleLogin,
	"register": handleRegister,
	"reset":    handleReset,
	"users":    handleUsers,
}

func (c *Command) Run(state *state.AppState) error {
	callback, ok := mapCommands[c.Name]

	if !ok {
		return errors.New("uknown command")
	}

	err := callback(state, c.Params)

	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	return nil
}

// Handlers

func handleLogin(state *state.AppState, params []string) error {
	if len(params) != 1 {
		return errors.New("login command expects 1 param : <username>")
	}

	username := params[0]

	_, err := state.Db.FindUserByName(
		context.Background(),
		username,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user doesn't exist in db, register first")
		}
		return fmt.Errorf("error finding user by name: %w", err)
	}

	err = state.Cfg.SetUser(username)

	if err != nil {
		return fmt.Errorf("err setting user: %w", err)
	}

	fmt.Printf("Current user is %s\n", username)

	return nil
}

func handleRegister(state *state.AppState, params []string) error {
	if len(params) != 1 {
		return errors.New("register command expects 1 param : <username>")
	}

	username := params[0]

	_, err := state.Db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID: uuid.NullUUID{
				UUID:  uuid.New(),
				Valid: true,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      username,
		},
	)

	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	err = state.Cfg.SetUser(username)

	if err != nil {
		return fmt.Errorf("err setting user: %w", err)
	}

	fmt.Printf("Current user is %s\n", username)

	return nil
}

func handleReset(state *state.AppState, params []string) error {
	err := state.Db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error del users: %w", err)
	}

	fmt.Println("All users have been deleted !")

	return nil
}

func handleUsers(state *state.AppState, params []string) error {
	users, err := state.Db.GetAllUsers(context.Background())

	if err != nil {
		return fmt.Errorf("err getting all users: %w", err)
	}

	for _, user := range users {
		str := "* " + user.Name

		if state.Cfg.Current_username == user.Name {
			str += " (current)"
		}

		fmt.Println(str)
	}

	return nil
}
