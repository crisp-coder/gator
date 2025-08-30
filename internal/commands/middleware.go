package commands

import (
	"context"
	"fmt"

	"github.com/crisp-coder/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {

		if s.Cfg.Username == "" {
			return fmt.Errorf("not logged in")
		}

		logged_in_user, err := s.Db.GetUserByName(context.Background(), s.Cfg.Username)
		if err != nil {
			return fmt.Errorf("error retrieving user data: %w", err)
		}
		return handler(s, cmd, logged_in_user)
	}
}
