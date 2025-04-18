package state

import (
	"database/sql"
	"fmt"

	"github.com/Ciobi0212/gator.git/internal/config"
	"github.com/Ciobi0212/gator.git/internal/database"
)

type AppState struct {
	Cfg *config.Config
	Db  *database.Queries
}

func GetInitState() (*AppState, error) {
	cfg, err := config.ReadConfig()

	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	db, err := sql.Open("postgres", cfg.Db_url)

	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %w", err)
	}

	dbQueries := database.New(db)

	state := AppState{
		Cfg: cfg,
		Db:  dbQueries,
	}

	return &state, nil
}
