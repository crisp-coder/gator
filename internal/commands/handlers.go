package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/crisp-coder/gator/internal/database"
	"github.com/crisp-coder/gator/internal/rss"
	"github.com/google/uuid"
)

func handlerHelp(s *State, cmd Command) error {
	fmt.Printf("help\n")
	fmt.Printf("login <username> - logs in the user.\n")
	fmt.Printf("register <username> - adds a user to the database and automatically logs in the user.\n")
	fmt.Printf("reset - drops rows data but keep tables.\n")
	fmt.Printf("users - lists all users.\n")
	fmt.Printf("feeds - lists all feeds.\n")
	fmt.Printf("addfeed <name> <url>\n")
	fmt.Printf("agg - print rss feeds to console.")
	fmt.Printf("follow <url> - adds the feed for the url to the users follows.\n")
	fmt.Printf("following - lists all feeds followed by the current user.\n")
	fmt.Printf("unfollow <url> - removes follow for url for current user.\n")
	fmt.Printf("browse <limit> - prints up to limit posts for user feeds.\n")
	return nil
}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("missing arguments to login command")
	}

	user_res, err := s.Db.GetUserByName(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error retrieving user from database: %w", err)
	}

	err = s.Cfg.SetUser(user_res.Name)
	if err != nil {
		return fmt.Errorf("error handling login: %w", err)
	}

	fmt.Printf("User: %v has been set.\n", s.Cfg.Username)
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 || cmd.Args[0] == "" {
		return errors.New("missing arguments to register command")
	}

	user_res, err := s.Db.CreateUser(
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

	err = s.Cfg.SetUser(user_res.Name)
	if err != nil {
		return fmt.Errorf("error handling register: %w", err)
	}

	fmt.Printf("User: %v has been set.\n", s.Cfg.Username)

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
	if len(cmd.Args) != 1 {
		return fmt.Errorf("missing time between requests argument to agg command")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error parsing time between requests: %w", err)
	}

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		err := ScrapeFeeds(s)
		if err != nil {
			return fmt.Errorf("error in scrape feeds: %w", err)
		}
	}
}

func ScrapeFeeds(s *State) error {
	// Get next feed in database
	feed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Update feed time in database
	_, err = s.Db.MarkFeedFetched(
		context.Background(),
		database.MarkFeedFetchedParams{
			ID:            feed.ID,
			LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Get rss feed data from provider url
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rss_feed, err := rss.FetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	layouts := []string{
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC822Z,  // "02 Jan 06 15:04 -0700"
		time.RFC822,   // "02 Jan 06 15:04 MST"
		time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
	}

	// Save each item in the rss feed to the posts table
	for _, item := range rss_feed.Channel.Item {
		// Try to parse publish date with all date layouts
		var parse_err error
		var published_at time.Time
		for _, layout := range layouts {
			t, err := time.Parse(layout, item.PubDate)
			if err == nil {
				published_at = t
				parse_err = nil
				break
			} else {
				parse_err = err
			}

		}
		// Log error and continue
		if parse_err != nil {
			fmt.Println(fmt.Errorf("%w", parse_err))
			fmt.Printf("error parsing publish date. skipping item %v...\n", item)
			continue
		}

		post, err := s.Db.CreatePost(
			context.Background(),
			database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Title:       item.Title,
				Url:         item.Link,
				Description: sql.NullString{String: item.Description, Valid: true},
				PublishedAt: published_at,
				FeedID:      feed.ID,
			})

		if err != nil {
			// Note that the create query fails on duplicate items, so
			// this normally gets printed a lot.
			//fmt.Printf("error saving post to database: %v", err)
			continue
		}
		fmt.Printf("Saved post: %v\n", post)
	}

	return nil
}

func handleAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("missing arguments to addfeed")
	}

	feedname := cmd.Args[0]
	url := cmd.Args[1]

	// Add the feed to the database
	feed_res, err := s.Db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      feedname,
			Url:       url,
			UserID:    user.ID,
		})

	if err != nil {
		return fmt.Errorf("error creating new feed")
	}

	// Automatically add feed follow for logged in user.
	follow_res, err := s.Db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    user.ID,
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

func handleFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("missing arguments for follow command")
	}

	url := cmd.Args[0]
	feed_res, err := s.Db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error getting feed by url: %w", err)
	}

	follow_res, err := s.Db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    user.ID,
			FeedID:    feed_res.ID,
		})

	if err != nil {
		return fmt.Errorf("error inserting feed follows: %w", err)
	}

	fmt.Printf("User: %v\n", follow_res.Username)
	fmt.Printf("Feed: %v\n", follow_res.Feedname)

	return nil
}

func handleFollowing(s *State, cmd Command, user database.User) error {
	feed_follows, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	fmt.Printf("User: %v\n", user.Name)
	for _, ff := range feed_follows {
		fmt.Printf("Feed: %v\n", ff.Feedname)
		fmt.Printf("URL: %v\n", ff.Url)
	}

	return nil
}

func handleUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("missing arguments for unfollow command")
	}
	url := cmd.Args[0]

	feed, err := s.Db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("deleting feed: %v", feed)

	err = s.Db.DeleteFeedFollow(context.Background(),
		database.DeleteFeedFollowParams{
			UserID: user.ID,
			FeedID: feed.ID,
		})

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func handleBrowse(s *State, cmd Command, user database.User) error {
	limit := int64(2)
	var err error
	if len(cmd.Args) == 1 {
		limit, err = strconv.ParseInt(cmd.Args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("error parsing limit: %w", err)
		}
	}

	posts, err := s.Db.GetPostsForUser(
		context.Background(),
		database.GetPostsForUserParams{
			UserID:  user.ID,
			Column2: limit,
		})

	if err != nil {
		return fmt.Errorf("error retrieving posts for user: %w", err)
	}

	for _, post := range posts {
		fmt.Printf("Title: %v\n", post.Title)
		fmt.Printf("Link: %v\n", post.Url)
		fmt.Printf("Description: %v\n", post.Description)
		fmt.Printf("PubDate: %v\n", post.PublishedAt)
	}

	return nil
}
