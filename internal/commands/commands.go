package commands

import (
	"errors"
	"fmt"
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
		fmt.Printf("command: %v\n", cmd.Name)
		for i, arg := range cmd.Args {
			fmt.Printf("arg[%v]: %v\n", i, arg)
		}
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

	cmds.Register("help", handlerHelp)
	cmds.Register("login", handlerLogin)
	cmds.Register("register", handlerRegister)
	cmds.Register("reset", handleReset)
	cmds.Register("users", handleListUsers)
	cmds.Register("agg", handleAgg)
	cmds.Register("addfeed", middlewareLoggedIn(handleAddFeed))
	cmds.Register("feeds", handleListFeeds)
	cmds.Register("follow", middlewareLoggedIn(handleFollow))
	cmds.Register("following", middlewareLoggedIn(handleFollowing))
	cmds.Register("unfollow", middlewareLoggedIn(handleUnfollow))

	return cmds
}
