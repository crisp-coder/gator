package commands

import (
	"errors"
	"fmt"
)

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("missing arguments to login command")
	}

	err := s.Cfg.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error handling login: %w", err)
	}

	fmt.Printf("User: %v has been set.\n", cmd.Args[0])
	return nil
}
