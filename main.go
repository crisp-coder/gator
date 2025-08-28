package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/crisp-coder/gator/internal/commands"
	"github.com/crisp-coder/gator/internal/config"
	"github.com/crisp-coder/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {

	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.Db_url)
	if err != nil {
		fmt.Println(err)
	}
	dbQueries := database.New(db)

	state := commands.State{
		Db:  dbQueries,
		Cfg: &cfg,
	}

	cmds := commands.MakeCommands()

	args := os.Args
	if len(args) < 2 {
		fmt.Println("missing command name")
		os.Exit(1)
	}
	cmd_name := args[1]
	cmd_args := args[2:]

	err = cmds.Run(&state, commands.Command{
		Name: cmd_name,
		Args: cmd_args,
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
