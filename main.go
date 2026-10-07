package main

import (
	"fmt"
	"gator/internal/config"
	"log"
	"os"
)

func main() {
	// need to change to a pointer
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	var st state
	st.conf = conf

	cmds := commands{
		cmdMap: make(map[string]func(st *state, cmd command) error),
	}
	cmds.register("login", handlerLogin)

	if len(os.Args) < 2 {
		fmt.Println("Usage: gator [command] [args]")
		os.Exit(1)
	}

	var thisCommand command
	thisCommand.name = os.Args[1]
	if len(os.Args) > 2 {
		thisCommand.args = os.Args[2:]
	}

	err = cmds.run(&st, thisCommand)
	if err != nil {
		log.Fatal(err)
	}

}
