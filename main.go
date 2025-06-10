package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/curtisbraxdale/blog-gator/internal/config"
	"github.com/curtisbraxdale/blog-gator/internal/database"
	"github.com/curtisbraxdale/blog-gator/internal/rss"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	value, exists := c.commandMap[cmd.name]
	if exists {
		err := value(s, cmd)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Command not found.")
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}

func main() {
	fig, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config file.")
		return
	}

	appState := state{config: &fig}
	db, err := sql.Open("postgres", fig.DbUrl)
	dbQueries := database.New(db)
	appState.db = dbQueries
	cliCommands := commands{make(map[string]func(*state, command) error)}
	cliCommands.register("login", handlerLogin)
	cliCommands.register("register", handlerRegister)
	cliCommands.register("reset", handlerReset)
	cliCommands.register("users", handlerUsers)
	cliCommands.register("agg", handlerAgg)
	cliCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cliCommands.register("feeds", handlerFeeds)
	cliCommands.register("follow", middlewareLoggedIn(handlerFollow))
	cliCommands.register("following", middlewareLoggedIn(handlerFollowing))
	cliCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cliCommands.register("browse", middlewareLoggedIn(handlerBrowse))

	cliArguments := os.Args
	if len(cliArguments) < 2 {
		err = fmt.Errorf("Not enough arguments.")
		fmt.Printf("Error Found: %v\n", err)
		os.Exit(1)
	}

	commandName := cliArguments[1]
	commandArguments := cliArguments[2:]
	newCommand := command{name: commandName, arguments: commandArguments}
	err = cliCommands.run(&appState, newCommand)
	if err != nil {
		fmt.Printf("Error Found: %v\n", err)
		os.Exit(1)
	}
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("No Arguments")
	}
	username := cmd.arguments[0]
	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		os.Exit(1)
	}
	err = s.config.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("User has been set to: %v", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("No Arguments")
	}
	username := cmd.arguments[0]
	userParams := database.CreateUserParams{ID: uuid.New(), CreatedAt: sql.NullTime{Time: time.Now(), Valid: true}, UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true}, Name: username}
	newUser, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		fmt.Println("User already exists.")
		os.Exit(1)
	}
	err = s.config.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Println("User Created:")
	fmt.Print(newUser)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetDB(context.Background())
	if err != nil {
		os.Exit(1)
	}
	fmt.Println("Succesfully reset table.")
	os.Exit(0)
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		os.Exit(1)
	}
	for _, user := range users {
		if user == s.config.CurrentUserName {
			fmt.Printf("* %v (current)\n", user)
		} else {
			fmt.Printf("* %v\n", user)
		}
	}
	os.Exit(0)
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("Not enough arguments.")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return errors.New("Not enough arguments.")
	}
	feed_params := database.CreateFeedParams{ID: uuid.New(), CreatedAt: sql.NullTime{Time: time.Now(), Valid: true}, UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true}, Name: cmd.arguments[0], Url: cmd.arguments[1], UserID: user.ID}
	new_feed, err := s.db.CreateFeed(context.Background(), feed_params)
	if err != nil {
		return err
	}

	feedFollowParams := database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: sql.NullTime{Time: time.Now(), Valid: true}, UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true}, UserID: user.ID, FeedID: new_feed.ID}
	_, err = s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return err
	}

	fmt.Println("Feed created:")
	fmt.Print(new_feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feedRows, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feedRow := range feedRows {
		user, err := s.db.GetUsername(context.Background(), feedRow.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("Name: %s\n", feedRow.Name)
		fmt.Printf("URL: %s\n", feedRow.Url)
		fmt.Printf("Username: %s\n\n", user.Name)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return errors.New("Not enough arguments.")
	}
	feed_id, err := s.db.GetFeedID(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}
	feedFollowParams := database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: sql.NullTime{Time: time.Now(), Valid: true}, UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true}, UserID: user.ID, FeedID: feed_id}
	newFeedFollows, err := s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return err
	}
	fmt.Printf("(User: %v) now follows (Feed: %v)\n", newFeedFollows.UserName, newFeedFollows.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	following, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	fmt.Printf("%v follows:\n", s.config.CurrentUserName)
	for _, followRow := range following {
		fmt.Printf("%v\n", followRow.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return errors.New("Not enough arguments.")
	}
	feed_url := cmd.arguments[0]
	feed_id, err := s.db.GetFeedID(context.Background(), feed_url)
	if err != nil {
		return err
	}
	deleteParams := database.DeleteFeedFollowParams{UserID: user.ID, FeedID: feed_id}
	err = s.db.DeleteFeedFollow(context.Background(), deleteParams)
	if err != nil {
		return err
	}
	return nil
}

func scrapeFeeds(s *state) error {
	next_feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(), next_feed.ID)
	if err != nil {
		return err
	}
	feed, err := rss.FetchFeed(context.Background(), next_feed.Url)
	if err != nil {
		return err
	}
	for _, item := range feed.Channel.Item {
		pub_date, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return err
		}
		post := database.CreatePostParams{ID: uuid.New(), CreatedAt: sql.NullTime{Time: time.Now(), Valid: true}, UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true}, Title: item.Title, Url: item.Link, Description: item.Description, PublishedAt: sql.NullTime{Time: pub_date, Valid: true}, FeedID: next_feed.ID}
		_, err = s.db.CreatePost(context.Background(), post)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			continue
		}
		if err != nil {
			log.Printf("failed to create post: %v", err)
		}
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := int32(2)
	if len(cmd.arguments) > 0 {
		limit64, err := strconv.ParseInt(cmd.arguments[0], 10, 32)
		if err != nil {
			return err
		}
		limit = int32(limit64)
	}
	get_post_params := database.GetPostsForUserParams{UserID: user.ID, Limit: int32(limit)}
	posts, err := s.db.GetPostsForUser(context.Background(), get_post_params)
	if err != nil {
		return err
	}
	fmt.Print("Browsing Posts\n\n")
	for _, post := range posts {
		fmt.Printf("\nTitle: %v\nDescription: %v\nPublished: %v\n", post.Title, post.Description, post.PublishedAt.Time.Format("2006-01-02"))
	}
	if len(posts) == 0 {
		fmt.Println("No posts found for your feeds!")
	}
	return nil
}
