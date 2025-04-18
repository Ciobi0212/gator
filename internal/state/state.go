package state

import (
	"fmt"

	"github.com/Ciobi0212/gator.git/internal/config"
)

type AppState struct {
	Cfg *config.Config
}

func GetInitState() (*AppState, error) {
	cfg, err := config.ReadConfig()

	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	state := AppState{
		Cfg: cfg,
	}

	return &state, nil
}
