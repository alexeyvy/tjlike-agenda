package repost

import (
	"github.com/alexeyvy/tjlike-agenda/domain"
	"testing"
	"time"
)

func TestInMemoryInsertAndRead(t *testing.T) {
	s := NewInMemoryStore()
	s.insert(&entry{
		R: domain.NewRepost(domain.NewPublication("platform/id", 10300, time.Now()), time.Now(), domain.SuggestionRate(5)),
	})
	s.insert(&entry{
		R: domain.NewRepost(domain.NewPublication("platform/id2", 10300, time.Now()), time.Now(), domain.SuggestionRate(5)),
	})

	entries := make([]entry, 0)
	s.walkAll(func(e *entry) {
		entries = append(entries, *e)
	})
	if len(entries) != 2 {
		t.Errorf(
			"Expected 2 entries in the store, got %d", len(entries),
		)
	}
}

func TestInMemoryIdSequence(t *testing.T) {
	s := NewInMemoryStore()
	e1 := entry{
		R: domain.NewRepost(domain.NewPublication("platform/id", 10300, time.Now()), time.Now(), domain.SuggestionRate(5)),
	}
	s.insert(&e1)
	e2 := entry{
		R: domain.NewRepost(domain.NewPublication("platform/id2", 10300, time.Now()), time.Now(), domain.SuggestionRate(5)),
	}
	s.insert(&e2)

	s.walkAll(func(e *entry) {
		if e.Id != 1 && e.Id != 2 {
			t.Errorf(
				"Expected sequential IDs 1 and 2, got %d", e.Id,
			)
		}
	})
}

func TestInMemoryDeletion(t *testing.T) {
	s := NewInMemoryStore()

	e1 := entry{
		R: domain.NewRepost(domain.NewPublication("platform/id", 10300, time.Now()), time.Now(), domain.SuggestionRate(5)),
	}
	s.insert(&e1)
	e2 := entry{
		R: domain.NewRepost(domain.NewPublication("platform/id2", 10300, time.Now()), time.Now(), domain.SuggestionRate(5)),
	}
	s.insert(&e2)

	s.delete(e1)

	entries := make([]entry, 0)
	s.walkAll(func(e *entry) {
		entries = append(entries, *e)
	})
	if len(entries) != 1 {
		t.Errorf(
			"Expected 1 entry left in the store after deleting 1, got %d", len(entries),
		)
	}

	s.delete(e2)

	entries = make([]entry, 0)
	s.walkAll(func(e *entry) {
		entries = append(entries, *e)
	})
	if len(entries) != 0 {
		t.Errorf(
			"Expected no entry left in the store after deleting 2, got %d", len(entries),
		)
	}
}
