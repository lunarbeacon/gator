package main

import (
	"context"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"
	"time"

	"github.com/google/uuid"
)

type state struct {
	db   *database.Queries
	conf *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmdMap map[string]func(st *state, cmd command) error
}

func (c *commands) run(st *state, cmd command) error {
	/*
		fmt.Println(os.Args)
		if len(cmd.args) < 2 {
			fmt.Println("In run()")
			fmt.Println("Usage: gator [command]")
			os.Exit(1)
		}
	*/

	f := c.cmdMap[cmd.name]
	err := f(st, cmd)
	if err != nil {
		os.Exit(1)
	}
	return nil
}

func (c *commands) register(name string, f func(st *state, cmd command) error) {
	c.cmdMap[name] = f
	return
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		fmt.Println("Usage: gator login <username>")
		return fmt.Errorf("username not supplied for %s", cmd.name)
	}
	// check if user exists
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Printf("User doesn't exist: %s\n", cmd.args[0])
		os.Exit(1)
	}

	err = config.SetUser(s.conf, cmd.args[0])
	if err != nil {
		return fmt.Errorf("could not set user %s: %v", cmd.name, err)
	}
	fmt.Println("User login set to ", cmd.args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		fmt.Println("Usage: gator register <username>")
		return fmt.Errorf("username not supplied for %s", cmd.name)
	}
	// add the user to the users table in the database after checking
	// whether it already exists
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		fmt.Printf("User already exists: %s\n", cmd.args[0])
		os.Exit(1)
	}
	parms := database.CreateUserParams{}
	parms.ID = uuid.New()
	parms.CreatedAt = time.Now()
	parms.UpdatedAt = time.Now()
	parms.Name = cmd.args[0]
	user, err := s.db.CreateUser(context.Background(), parms)
	if err != nil {
		return fmt.Errorf("could not create user %s: %v", cmd.args[0], err)
	}

	// set the username in the config file
	err = config.SetUser(s.conf, user.Name)
	if err != nil {
		return fmt.Errorf("could not set user %s: %v", user.Name, err)
	}
	fmt.Println("User login set to ", cmd.args[0])
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not get users: %v", err)
	}
	currentUser := s.conf.CurrentUserName

	for _, user := range users {
		if user.Name == currentUser {
			fmt.Println(user.Name, "(current)")
		} else {
			fmt.Println(user.Name)
		}
	}
	return nil
}

func handlerFetchFeed(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		fmt.Println("Usage: gator agg <time between requests> <url>")
		return fmt.Errorf("no time between requests provided: %s", cmd.name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		fmt.Printf("Invalid time between requests: %s\n", cmd.args[0])
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

}

func handlerAddFeed(s *state, cmd command, user database.User) error {

	if len(cmd.args) != 2 {
		fmt.Println("Usage: gator addfeed <feed name> <url>")
		return fmt.Errorf("feed name and url not supplied for %s", cmd.name)
	}

	// get the user id of the current user
	// user, err := s.db.GetUser(context.Background(), s.conf.CurrentUserName)
	//if err != nil {
	//	return fmt.Errorf("could not get user %s: %v", s.conf.CurrentUserName, err)
	//}
	//userID := user.ID

	// create the feed in the database
	parms := database.CreateFeedParams{}
	parms.ID = uuid.New()
	parms.CreatedAt = time.Now()
	parms.UpdatedAt = time.Now()
	parms.Name = cmd.args[0]
	parms.Url = cmd.args[1]
	parms.UserID = user.ID
	feed, err := s.db.CreateFeed(context.Background(), parms)
	if err != nil {
		return fmt.Errorf("could not create feed %s: %v", cmd.args[0], err)
	}

	// follow the feed
	followparms := database.CreateFeedFollowParams{}
	followparms.ID = uuid.New()
	followparms.CreatedAt = time.Now()
	followparms.UpdatedAt = time.Now()
	followparms.UserID = user.ID
	followparms.FeedID = feed.ID
	_, err = s.db.CreateFeedFollow(context.Background(), followparms)
	if err != nil {
		return fmt.Errorf("could not follow feed %s: %v", cmd.args[0], err)
	}

	fmt.Printf("New feed created and followed: %s\n", feed.Name)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("could not get feeds: %v", err)
	}

	for _, feed := range feeds {
		fmt.Printf("%s\t\t %s\t\t %s\n", feed.Name, feed.Url, feed.Username)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		fmt.Println("Usage: gator follow <feed url>")
		return fmt.Errorf("feed url not supplied for %s", cmd.name)
	}

	// get the feed_id of the feed to follow
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Printf("Feed not found: %v\n", err)
		return err
	}
	feedID := feed.ID

	parms := database.CreateFeedFollowParams{}
	parms.ID = uuid.New()
	parms.CreatedAt = time.Now()
	parms.UpdatedAt = time.Now()
	parms.UserID = user.ID
	parms.FeedID = feedID
	feedfollow, err := s.db.CreateFeedFollow(context.Background(), parms)
	if err != nil {
		return fmt.Errorf("could not follow feed %s: %v", cmd.args[0], err)
	}

	fmt.Printf("%s followed feed: %s\n", feedfollow.UserName, feedfollow.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		fmt.Printf("No feeds being followed: %v\n", err)
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("%s\n", feed.Feedname)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		fmt.Println("Usage: gator unfollow <feed name>")
		return nil
	}

	// get the feed ID using the URL passed as an arg
	feedURL := cmd.args[0]
	rssFeed, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		fmt.Printf("Feed not found %s: %v\n", feedURL, err)
		return err
	}
	feedID := rssFeed.ID
	parms := database.DeleteFeedFollowParams{}
	parms.UserID = user.ID
	parms.FeedID = feedID

	err = s.db.DeleteFeedFollow(context.Background(), parms)
	if err != nil {
		fmt.Printf("Could not delete %s: %v\n", feedURL, err)
		return err
	}

	return nil
}

func handlerReset(s *state, cmd command) error {

	// user apocalypse
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Reset failed: %v\n", err)
	}

	return nil
}
