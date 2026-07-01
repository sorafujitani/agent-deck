package deck

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreRoundTrip(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "deck.json"))
	deck := NewDeck()
	if _, err := deck.AddTask(NewTaskInput{Goal: "write MVP"}, time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	if err := store.Save(deck); err != nil {
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
