package deck

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const storeFileName = "deck.json"

type Store struct {
	path string
}

func NewStore(path string) Store {
	return Store{path: path}
}

func DefaultStore() (Store, error) {
	if root := os.Getenv("AGENT_DECK_HOME"); root != "" {
		return Store{path: filepath.Join(root, storeFileName)}, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return Store{}, fmt.Errorf("resolve config dir: %w", err)
	}
	return Store{path: filepath.Join(dir, "agent-deck", storeFileName)}, nil
}

func (s Store) Path() string {
	return s.path
}

func (s Store) Load() (Deck, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return NewDeck(), nil
	}
	if err != nil {
		return Deck{}, fmt.Errorf("read store: %w", err)
	}
	if len(data) == 0 {
		return NewDeck(), nil
	}

	var deck Deck
	if err := json.Unmarshal(data, &deck); err != nil {
		return Deck{}, fmt.Errorf("decode store: %w", err)
	}
	if deck.Version == 0 {
		deck.Version = 1
	}
	if deck.Tasks == nil {
		deck.Tasks = []Task{}
	}
	return deck, nil
}

func (s Store) Save(deck Deck) error {
	if deck.Version == 0 {
		deck.Version = 1
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create store dir: %w", err)
	}
	data, err := json.MarshalIndent(deck, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write store: %w", err)
	}
	return nil
}
