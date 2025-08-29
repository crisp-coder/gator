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

	res, err := s.Db.GetUserByName(context.Background(), cmd.Args[0])
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

	users_res, err := s.Db.ListUsers(context.Background())

	if err != nil {
		return fmt.Errorf("filed to list users: %w", err)
	}

	for _, val := range users_res {
		fmt.Printf(" - %v", val.Name)
		if val.Name == s.Cfg.Username {
			fmt.Printf(" (current)\n")
		} else {
			fmt.Printf("\n")
		}
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

	user_res, err := s.Db.GetUserByName(context.Background(), s.Cfg.Username)
	if err != nil {
		return fmt.Errorf("error retrieving user from database: %w", err)
	}

	name := cmd.Args[0]
	url := cmd.Args[1]
	userId := user_res.ID

	feed_res, err := s.Db.CreateFeed(
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

	fmt.Printf("For User ID: %v\n", feed_res.UserID)
	fmt.Printf("Feedname: %v\n", feed_res.Name)
	fmt.Printf("URL: %v\n", feed_res.Url)

	// Automatically add feed follow for logged in user.
	follow_res, err := s.Db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    user_res.ID,
			FeedID:    feed_res.ID,
		})

	if err != nil {
		return fmt.Errorf("error inserting feed follows: %w", err)
	}

	fmt.Printf("User: %v\n", follow_res.Username)
	fmt.Printf("Feed: %v\n", follow_res.Feedname)

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
		if feed.Username.Valid {
			fmt.Printf("Username: %v\n", feed.Username.String)
		}
		fmt.Printf("Feed: %v\n", feed.Name)
		fmt.Printf("URL: %v\n", feed.Url)
	}

	return nil
}

func handleFollow(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("missing arguments to follow command")
	}

	url := cmd.Args[0]
	feed_res, err := s.Db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error getting feed by url: %w", err)
	}

	user_res, err := s.Db.GetUserByName(context.Background(), s.Cfg.Username)
	if err != nil {
		return fmt.Errorf("error getting user by name: %w", err)
	}

	follow_res, err := s.Db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    user_res.ID,
			FeedID:    feed_res.ID,
		})

	if err != nil {
		return fmt.Errorf("error inserting feed follows: %w", err)
	}

	fmt.Printf("User: %v\n", follow_res.Username)
	fmt.Printf("Feed: %v\n", follow_res.Feedname)

	return nil
}

func handleFollowing(s *State, cmd Command) error {
	fmt.Printf("Querying id for user: %v\n", s.Cfg.Username)
	users_res, err := s.Db.GetUserByName(context.Background(), s.Cfg.Username)
	if err != nil {
		return fmt.Errorf("%W", err)
	}

	fmt.Printf("found id: %v\n", users_res.ID)
	fmt.Printf("Querying feed follows:\n")
	feed_follows, err := s.Db.GetFeedFollowsForUser(context.Background(), users_res.ID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("Found %v follows:\n", len(feed_follows))
	fmt.Printf("User: %v\n", users_res.Name)
	for _, ff := range feed_follows {
		fmt.Printf("Feed: %v\n", ff.Feedname)
		fmt.Printf("URL: %v\n", ff.Url)
	}

	return nil
}
