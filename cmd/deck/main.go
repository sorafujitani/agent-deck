package main

import (
	"os"

	"github.com/sorafujitani/agent-deck/internal/deck"
)

func main() {
	os.Exit(deck.Main(os.Args[1:], os.Stdout, os.Stderr))
}
