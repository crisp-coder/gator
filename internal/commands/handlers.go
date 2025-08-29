package commands

import (
	"context"
	"errors"
	"fmt"
	"html"
	"os"
	"os/signal"
	"time"

	"github.com/crisp-coder/gator/internal/database"
	"github.com/crisp-coder/gator/internal/rss"
	"github.com/google/uuid"
)

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("missing arguments to login command")
	}

	res, err := s.Db.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error retrieving user from database: %w", err)
	}

	err = s.Cfg.SetUser(res.Name)
	if err != nil {
		return fmt.Errorf("error handling login: %w", err)
	}

	fmt.Printf("User: %v has been set.\n", res.Name)
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 || cmd.Args[0] == "" {
		return errors.New("missing arguments to register command")
	}

	res, err := s.Db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.Args[0],
		})

	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	err = s.Cfg.SetUser(res.Name)
	if err != nil {
		return fmt.Errorf("error handling register: %w", err)
	}

	fmt.Printf("User: %v has been set.\n", res.Name)

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

func handleListUsers(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("too many arguments to list users command")
	}

	res, err := s.Db.ListUsers(context.Background())

	if err != nil {
		return fmt.Errorf("filed to list users: %w", err)
	}

	for _, val := range res {
		fmt.Printf(" - %v", val.Name)
		if val.Name == s.Cfg.Username {
			fmt.Printf(" (current)\n")
		}
		fmt.Printf("\n")
	}

	return nil
}

func handleAgg(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("missing arguments to agg command")
	}

	url := cmd.Args[0]

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	feed, err := rss.FetchFeed(ctx, url)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	fmt.Println(html.UnescapeString(feed.Channel.Title))
	fmt.Println(html.UnescapeString(feed.Channel.Link))
	fmt.Println(html.UnescapeString(feed.Channel.Description))
	indent := "  "
	for _, item := range feed.Channel.Item {
		fmt.Println(indent + html.UnescapeString(item.Title))
		fmt.Println(indent + html.UnescapeString(item.Link))
		fmt.Println(indent + html.UnescapeString(item.Description))
		fmt.Println(indent + html.UnescapeString(item.PubDate))
	}
	return nil
}

func handleAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("missing arguments to addfeed")
	}

	userres, err := s.Db.GetUser(context.Background(), s.Cfg.Username)
	if err != nil {
		return fmt.Errorf("error retrieving user from database: %w", err)
	}

	name := cmd.Args[0]
	url := cmd.Args[1]
	userId := userres.ID

	feedres, err := s.Db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      name,
			Url:       url,
			UserID:    userId,
		})
	if err != nil {
		return fmt.Errorf("error creating new feed")
	}

	fmt.Println(feedres)

	return nil
}

func handleListFeeds(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("too many arguments to feeds")
	}

	feeds, err := s.Db.ListFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("Name: %v\n", feed.Name)
		fmt.Printf("URL: %v\n", feed.Url)
		if feed.Username.Valid {
			fmt.Printf("Username: %v\n", feed.Username.String)
		}
	}

	return nil
}

func handleFollow(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("missing arguments to follow command")
	}

	return nil
}

func handleFollowing(s *State, cmd Command) error {
	return nil
}
