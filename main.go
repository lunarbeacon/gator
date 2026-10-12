package main

import (
	"context"
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// load the configuration file
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	var st state
	st.conf = conf

	// open a connection to the database
	db, err := sql.Open("postgres", conf.DbURL)
	dbQueries := database.New(db)
	st.db = dbQueries

	cmds := commands{
		cmdMap: make(map[string]func(st *state, cmd command) error),
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("users", handlerUsers)
	cmds.register("reset", handlerReset)
	cmds.register("agg", handlerFetchFeed)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	// if len(os.Args) < 2 {
	//	fmt.Println("Usage: gator [command] [args]")
	//	os.Exit(1)
	// }

	var thisCommand command
	thisCommand.name = os.Args[1]
	if len(os.Args) > 2 {
		thisCommand.args = os.Args[2:]
	}

	err = cmds.run(&st, thisCommand)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.conf.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}
