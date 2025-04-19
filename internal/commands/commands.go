package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Ciobi0212/gator.git/internal/database"
	"github.com/Ciobi0212/gator.git/internal/requests"
	"github.com/Ciobi0212/gator.git/internal/state"

	"github.com/google/uuid"
)

// Command names constants
const (
	CmdLogin     = "login"
	CmdRegister  = "register"
	CmdReset     = "reset"
	CmdUsers     = "users"
	CmdAgg       = "agg"
	CmdAddFeed   = "addfeed"
	CmdFeeds     = "feeds"
	CmdFollow    = "follow"
	CmdFollowing = "following"
	CmdUnfollow  = "unfollow"
	CmdBrowse    = "browse"
	CmdHelp      = "help"
)

type Command struct {
	Name   string
	Params []string
}

var mapCommands = make(map[string]func(*state.AppState, []string) error)

func registerCommand(name string, fun func(*state.AppState, []string) error) {
	if _, exists := mapCommands[name]; exists {
		log.Printf("Warning: Command '%s' is being registered more than once.", name)
	}
	mapCommands[name] = fun
}

func InitMapCommand() {
	registerCommand(CmdLogin, handleLogin)
	registerCommand(CmdRegister, handleRegister)
	registerCommand(CmdReset, handleReset)
	registerCommand(CmdUsers, handleUsers)
	registerCommand(CmdAgg, handleAgg)
	registerCommand(CmdAddFeed, middlewareLoggedIn(handleAddfeed))
	registerCommand(CmdFeeds, handleFeeds)
	registerCommand(CmdFollow, middlewareLoggedIn(handleFollow))
	registerCommand(CmdFollowing, middlewareLoggedIn(handleFollowing))
	registerCommand(CmdUnfollow, middlewareLoggedIn(handleUnfollow))
	registerCommand(CmdBrowse, middlewareLoggedIn(handleBrowse))
	registerCommand(CmdHelp, handleHelp)
}

func (c *Command) Run(state *state.AppState) error {
	callback, ok := mapCommands[c.Name]

	if !ok {
		return fmt.Errorf("unknown command: %s", c.Name)
	}

	err := callback(state, c.Params)

	if err != nil {
		return fmt.Errorf("error running command '%s': %w", c.Name, err)
	}

	return nil
}

// Middleware
func middlewareLoggedIn(handler func(*state.AppState, []string, database.User) error) func(*state.AppState, []string) error {
	return func(state *state.AppState, params []string) error {
		user, err := state.Db.FindUserByName(context.Background(), state.Cfg.Current_username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("no such user exists, register and try again")
			}
			return fmt.Errorf("err exec query find user by name: %w", err)
		}

		return handler(state, params, user)
	}
}

// Handlers

func handleLogin(state *state.AppState, params []string) error {
	if len(params) != 1 {
		return errors.New("login command expects 1 param : <username>")
	}

	username := params[0]

	_, err := state.Db.FindUserByName(
		context.Background(),
		username,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user doesn't exist in db, register first")
		}
		return fmt.Errorf("error finding user by name: %w", err)
	}

	err = state.Cfg.SetUser(username)

	if err != nil {
		return fmt.Errorf("err setting user: %w", err)
	}

	fmt.Printf("Current user is %s\n", username)

	return nil
}

func handleRegister(state *state.AppState, params []string) error {
	if len(params) != 1 {
		return errors.New("register command expects 1 param : <username>")
	}

	username := params[0]

	_, err := state.Db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      username,
		},
	)

	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	err = state.Cfg.SetUser(username)

	if err != nil {
		return fmt.Errorf("err setting user: %w", err)
	}

	fmt.Printf("Current user is %s\n", username)

	return nil
}

func handleReset(state *state.AppState, params []string) error {
	err := state.Db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error del users: %w", err)
	}

	fmt.Println("All users have been deleted !")

	err = state.Db.DeleteAllFeeds(context.Background())

	if err != nil {
		return fmt.Errorf("error del feeds: %w", err)
	}

	fmt.Println("All feeds have been deleted !")

	return nil
}

func handleUsers(state *state.AppState, params []string) error {
	users, err := state.Db.GetAllUsers(context.Background())

	if err != nil {
		return fmt.Errorf("err getting all users: %w", err)
	}

	for _, user := range users {
		str := "* " + user.Name

		if state.Cfg.Current_username == user.Name {
			str += " (current)"
		}

		fmt.Println(str)
	}

	return nil
}

func handleAgg(state *state.AppState, params []string) error {
	if len(params) != 1 {
		return errors.New("agg command params : <timeBeetweenRequests>")
	}

	timeBetweenRequests, err := time.ParseDuration(params[0])
	if err != nil {
		return errors.New("invalid input format, example: agg 1s, agg 1m, agg 1h")
	}

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		feed, err := state.Db.GetNextFeedToFetch(context.Background())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No more feeds to fetch")
				break
			}

			fmt.Println(fmt.Errorf("error getting next feed to fetch: %w", err))
			break
		}

		fmt.Println("-----------" + feed.Name + "-----------")

		rssfeed, err := requests.FetchFeed(context.Background(), feed.Url)
		if err != nil {
			fmt.Println(fmt.Errorf("error fetching feed: %w", err))
			break
		}

		for _, item := range rssfeed.Items {
			fmt.Println(item.Title)

			publishedAt, err := time.Parse(time.RFC1123, item.Published)
			if err != nil {
				fmt.Printf("error parsing PubDate: %v\n", err)
				publishedAt = time.Time{}
			}

			// URL is unique, so if the post already is in the DB it will simply not inserted (see posts.sql)
			_, err = state.Db.CreatePost(
				context.Background(),
				database.CreatePostParams{
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					Title:       item.Title,
					Url:         item.Link,
					PublishedAt: publishedAt,
					Description: item.Description,
					FeedID:      feed.ID,
				},
			)

			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("err creating post: %w", err)
			}
		}

		err = state.Db.MarkFeedFetched(context.Background(), feed.ID)
		if err != nil {
			fmt.Println(fmt.Errorf("error marking feed fetched: %w", err))
			break
		}

		fmt.Println("----------------------")
	}

	return nil
}

func handleAddfeed(state *state.AppState, params []string, user database.User) error {
	if len(params) != 2 {
		return errors.New("addfeed command needs 2 params: <name> <url>")
	}

	name, url := params[0], params[1]

	feed, err := state.Db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			Name:      name,
			Url:       url,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("err creating feed: %w", err)
	}

	createFeedFollowRow, err := state.Db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			UserID:    user.ID,
			FeedID:    feed.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("err creating feed_follow entry: %w", err)
	}

	fmt.Printf("User %s now follows %s\n", createFeedFollowRow.Name, createFeedFollowRow.Name_2)

	fmt.Printf("%v\n", feed)

	return nil
}

func handleFeeds(state *state.AppState, params []string) error {
	if len(params) != 0 {
		return errors.New("no params required for feeds command")
	}

	feeds, err := state.Db.GetAllFeeds(context.Background())

	if err != nil {
		return fmt.Errorf("err getting all feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("Feed Name: %s\nURL: %s\n", feed.Name, feed.Url)
		fmt.Println("--------------")
	}

	return nil
}

func handleFollow(state *state.AppState, params []string, user database.User) error {
	if len(params) != 1 {
		return errors.New("follow commands needs 1 param: <url>")
	}

	url := params[0]

	feed, err := state.Db.FindFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("err follow, can't find feed: %w", err)
	}

	createFeedFollowRow, err := state.Db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			UserID:    user.ID,
			FeedID:    feed.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("err creating feed_follow entry: %w", err)
	}

	fmt.Printf("User %s now follows %s\n", createFeedFollowRow.Name, createFeedFollowRow.Name_2)

	return nil
}

func handleFollowing(state *state.AppState, params []string, user database.User) error {
	if len(params) != 0 {
		return errors.New("no params required for following command")
	}

	feedNames, err := state.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("err getting feeds for user: %w", err)
	}

	for _, name := range feedNames {
		fmt.Println(name)
	}

	return nil
}

func handleUnfollow(state *state.AppState, params []string, user database.User) error {
	if len(params) != 1 {
		return errors.New("unfollow command needs 1 param: <url>")
	}

	url := params[0]

	feed, err := state.Db.FindFeedByURL(context.Background(), url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no feed with specified url exists")
		}

		return fmt.Errorf("err query findFeedByUrl: %w", err)
	}

	err = state.Db.DeleteFeedFollowsEntry(
		context.Background(),
		database.DeleteFeedFollowsEntryParams{
			UserID: user.ID,
			FeedID: feed.ID,
		},
	)

	if err != nil {
		return fmt.Errorf("err query deleteFeedFollowsEntry: %w", err)
	}

	return nil
}

func handleBrowse(state *state.AppState, params []string, user database.User) error {
	if len(params) != 1 {
		return fmt.Errorf("browse command accepts 1 param: <numOfPosts>")
	}

	limit, err := strconv.Atoi(params[0])

	if err != nil {
		return errors.New("invalid input, it should be a number")
	}

	posts, err := state.Db.GetPostsForUser(
		context.Background(),
		database.GetPostsForUserParams{
			UserID: user.ID,
			Limit:  int32(limit),
		},
	)

	if err != nil {
		return fmt.Errorf("err getting posts for user: %w", err)
	}

	for _, post := range posts {
		fmt.Println("------------")
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("------------")
	}

	return nil
}

func handleHelp(state *state.AppState, params []string) error {
	fmt.Println("Gator - RSS Feed Aggregator")
	fmt.Println("===========================")
	fmt.Println()

	// App description
	fmt.Println("ABOUT:")
	fmt.Println("  Gator is a command-line RSS feed aggregator that helps you follow")
	fmt.Println("  and discover content from your favorite websites using RSS feeds.")
	fmt.Println("  It allows you to register as a user, subscribe to multiple feeds,")
	fmt.Println("  and browse the latest posts all from your terminal.")
	fmt.Println()

	fmt.Println("HOW IT WORKS:")
	fmt.Println("  1. Register or login to your account")
	fmt.Println("  2. Add or follow RSS feeds you're interested in")
	fmt.Println("  3. Start the aggregator to fetch the latest content")
	fmt.Println("  4. Browse posts from your followed feeds")
	fmt.Println()

	fmt.Println("WORKFLOW EXAMPLE:")
	fmt.Println("  ./gator register john         # Create a user account")
	fmt.Println("  ./gator addfeed 'Tech News' https://example.com/rss  # Add a feed")
	fmt.Println("  ./gator agg 10m &             # Start aggregation in background (every 10 min)")
	fmt.Println("  ./gator browse 20             # View the 20 most recent posts")
	fmt.Println()

	fmt.Println("AVAILABLE COMMANDS:")
	fmt.Println()

	// User management commands
	fmt.Println("User Management:")
	fmt.Println("  register <username>       - Create a new user account and login")
	fmt.Println("  login <username>          - Login as an existing user")
	fmt.Println("  users                     - List all registered users")

	// Feed management commands
	fmt.Println()
	fmt.Println("Feed Management:")
	fmt.Println("  addfeed <name> <url>      - Add a new RSS feed and follow it (requires login)")
	fmt.Println("  feeds                     - List all available feeds")
	fmt.Println("  follow <url>              - Follow an existing feed (requires login)")
	fmt.Println("  following                 - List all feeds you're following (requires login)")
	fmt.Println("  unfollow <url>            - Unfollow a feed (requires login)")

	// Content viewing commands
	fmt.Println()
	fmt.Println("Content:")
	fmt.Println("  browse <limit>            - View posts from feeds you follow (requires login)")
	fmt.Println("                              limit: number of posts to display")

	// System commands
	fmt.Println()
	fmt.Println("System:")
	fmt.Println("  agg <interval>            - Start feed aggregation process")
	fmt.Println("                              interval: time between fetches (e.g., 1s, 1m, 1h)")
	fmt.Println("  reset                     - Delete all users and feeds (use with caution)")
	fmt.Println("  help                      - Display this help information")

	// Examples
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ./gator register john     - Create a user named 'john'")
	fmt.Println("  ./gator addfeed 'Boot.dev Blog' https://blog.boot.dev/index.xml")
	fmt.Println("  ./gator agg 30s           - Fetch new content every 30 seconds")
	fmt.Println("  ./gator browse 10         - Show the 10 most recent posts")

	return nil
}
