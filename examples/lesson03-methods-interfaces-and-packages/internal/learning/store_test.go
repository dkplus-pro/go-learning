package learning

import (
	"errors"
	"testing"
)

func TestMemoryStoreSaveAndFindByName(t *testing.T) {
	store := NewMemoryStore()
	learner := Learner{Name: "Ada", CompletedLessons: 4}

	if err := store.Save(learner); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	got, err := store.FindByName("Ada")
	if err != nil {
		t.Fatalf("FindByName() unexpected error: %v", err)
	}

	if got != learner {
		t.Fatalf("FindByName() = %+v, want %+v", got, learner)
	}
}

func TestMemoryStoreSaveRequiresName(t *testing.T) {
	store := NewMemoryStore()

	err := store.Save(Learner{})

	if !errors.Is(err, ErrEmptyName) {
		t.Fatalf("Save() error = %v, want ErrEmptyName", err)
	}
}

func TestMemoryStoreFindByNameNotFound(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.FindByName("missing")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("FindByName() error = %v, want ErrNotFound", err)
	}
}
