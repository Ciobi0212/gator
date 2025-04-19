package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Ciobi0212/gator.git/internal/database"
	"github.com/Ciobi0212/gator.git/internal/requests"
	"github.com/Ciobi0212/gator.git/internal/state"

	"github.com/google/uuid"
)

type Command struct {
	Name   string
	Params []string
}

var mapCommands = map[string]func(*state.AppState, []string) error{
	"login":     handleLogin,
	"register":  handleRegister,
	"reset":     handleReset,
	"users":     handleUsers,
	"agg":       handleAgg,
	"addfeed":   middlewareLoggedIn(handleAddfeed),
	"feeds":     handleFeeds,
	"follow":    middlewareLoggedIn(handleFollow),
	"following": middlewareLoggedIn(handleFollowing),
	"unfollow":  middlewareLoggedIn(handleUnfollow),
	"browse":    middlewareLoggedIn(handleBrowse),
}

func registerCommand(name string, fun func(*state.AppState, []string) error) {
	mapCommands[name] = fun
}

func InitMapCommand() {
	registerCommand("login", handleLogin)
	registerCommand("register", handleRegister)
	registerCommand("reset", handleReset)
	registerCommand("users", handleUsers)
	registerCommand("agg", handleAgg)
	registerCommand("addfeed", middlewareLoggedIn(handleAddfeed))
	registerCommand("feeds", handleFeeds)
	registerCommand("follow", middlewareLoggedIn(handleFollow))
	registerCommand("following", middlewareLoggedIn(handleFollowing))
	registerCommand("unfollow", middlewareLoggedIn(handleUnfollow))
}

func (c *Command) Run(state *state.AppState) error {
	callback, ok := mapCommands[c.Name]

	if !ok {
		return errors.New("unknown command")
	}

	err := callback(state, c.Params)

	if err != nil {
		return fmt.Errorf("error running command: %w", err)
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
			fmt.Println(fmt.Errorf("error getting next feed to fetch: %w", err))
			break
		}

		fmt.Println("-----------" + feed.Name + "-----------")

		rssfeed, err := requests.FetchFeed(context.Background(), feed.Url)
		if err != nil {
			fmt.Println(fmt.Errorf("error fetching feed: %w", err))
			break
		}

		for _, item := range rssfeed.Channel.Item {
			fmt.Println(item.Title)

			publishedAt, err := time.Parse(time.RFC1123, item.PubDate)
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

			if err != nil {
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
