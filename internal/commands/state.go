package commands

import (
	"github.com/crisp-coder/gator/internal/config"
	"github.com/crisp-coder/gator/internal/database"
)

type State struct {
	Db  *database.Queries
	Cfg *config.Config
}
