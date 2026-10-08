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
