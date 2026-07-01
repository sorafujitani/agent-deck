package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sorafujitani/agent-deck/internal/deck"
)

const storeFileName = "deck.json"

type JSONStore struct {
	path string
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

func DefaultStore() (*JSONStore, error) {
	if root := os.Getenv("AGENT_DECK_HOME"); root != "" {
		return NewJSONStore(filepath.Join(root, storeFileName)), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve config dir: %w", err)
	}
	return NewJSONStore(filepath.Join(dir, "agent-deck", storeFileName)), nil
}

func (s *JSONStore) Path() string {
	return s.path
}

func (s *JSONStore) Load() (deck.Deck, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return deck.NewDeck(), nil
	}
	if err != nil {
		return deck.Deck{}, fmt.Errorf("read store: %w", err)
	}
	if len(data) == 0 {
		return deck.NewDeck(), nil
	}

	var state deck.Deck
	if err := json.Unmarshal(data, &state); err != nil {
		return deck.Deck{}, fmt.Errorf("decode store: %w", err)
	}
	if state.Version == 0 {
		state.Version = 1
	}
	if state.Tasks == nil {
		state.Tasks = []deck.Task{}
	}
	return state, nil
}

func (s *JSONStore) Save(state deck.Deck) error {
	if state.Version == 0 {
		state.Version = 1
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create store dir: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write store: %w", err)
	}
	return nil
}
