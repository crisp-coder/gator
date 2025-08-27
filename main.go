package main

import (
	"fmt"
	"os"

	"github.com/crisp-coder/gator/internal/commands"
	"github.com/crisp-coder/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("%w", err)
		os.Exit(1)
	}
	state := commands.State{
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
