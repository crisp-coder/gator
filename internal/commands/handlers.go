package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/crisp-coder/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("missing arguments to login command")
	}

	res, err := s.Db.GetUser(context.Background(), cmd.Args[0])
	fmt.Println(res)
	if err != nil {
		return fmt.Errorf("error retrieving user form database: %w", err)
	}

	err = s.Cfg.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error handling login: %w", err)
	}

	fmt.Printf("User: %v has been set.\n", cmd.Args[0])
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 || cmd.Args[0] == "" {
		return errors.New("missing arguments to register command")
	}

	res, err := s.Db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID: uuid.NullUUID{
				UUID:  uuid.New(),
				Valid: true,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.Args[0],
		})
	fmt.Printf("%v", res)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	err = s.Cfg.SetUser(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error handling register: %w", err)
	}

	fmt.Printf("User: %v has been set.\n", cmd.Args[0])

	return nil
}

func handleReset(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return errors.New("too many arguments to reset command")
	}

	err := s.Db.Reset(context.Background())

	if err != nil {
		return fmt.Errorf("failed to reset users: %w", err)
	}

	return nil
}
