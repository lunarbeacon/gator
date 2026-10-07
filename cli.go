package main

import (
	"fmt"
	"gator/internal/config"
	"os"
)

type state struct {
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
	err := config.SetUser(s.conf, cmd.args[0])
	if err != nil {
		return fmt.Errorf("could not set user %s: %v", cmd.name, err)
	}
	fmt.Println("User set to", cmd.name)
	return nil
}
