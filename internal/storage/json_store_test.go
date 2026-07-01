package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sorafujitani/agent-deck/internal/deck"
)

func TestJSONStoreRoundTrip(t *testing.T) {
	store := NewJSONStore(filepath.Join(t.TempDir(), "deck.json"))
	state := deck.NewDeck()
	if _, err := state.AddTask(
		"tsk_1",
		deck.NewTaskInput{Goal: "write MVP"},
		time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
	); err != nil {
		t.Fatal(err)
	}

	if err := store.Save(state); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Tasks) != 1 {
		t.Fatalf("tasks = %d, want 1", len(loaded.Tasks))
	}
	if loaded.Tasks[0].Goal != "write MVP" {
		t.Fatalf("goal = %s", loaded.Tasks[0].Goal)
	}
}
