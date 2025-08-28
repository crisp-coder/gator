package commands

import (
	"errors"
)

type Commands struct {
	cmd_map map[string]func(*State, Command) error
}

type Command struct {
	Name string
	Args []string
}

func (cmds *Commands) Run(s *State, cmd Command) error {
	if f, ok := cmds.cmd_map[cmd.Name]; ok {
		return f(s, cmd)
	} else {
		return errors.New("command not found")
	}
}

func (cmds *Commands) Register(name string, f func(*State, Command) error) {
	cmds.cmd_map[name] = f
}

func MakeCommands() Commands {
	cmds := Commands{
		cmd_map: make(map[string]func(*State, Command) error),
	}

	cmds.Register("login", handlerLogin)
	cmds.Register("register", handlerRegister)
	cmds.Register("reset", handleReset)
	cmds.Register("users", handleListUsers)
	cmds.Register("agg", handleAgg)

	return cmds
}
