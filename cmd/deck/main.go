package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sorafujitani/agent-deck/internal/cli"
	"github.com/sorafujitani/agent-deck/internal/deck"
	"github.com/sorafujitani/agent-deck/internal/storage"
)

func main() {
	store, err := storage.DefaultStore()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	service := deck.NewService(
		store,
		deck.WithClock(time.Now),
	)
	app := cli.NewApp(service, store.Path())
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
