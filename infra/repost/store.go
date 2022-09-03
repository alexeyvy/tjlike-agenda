package repost

import (
	"github.com/alexeyvy/tjlike-agenda/domain"
)

type entry struct {
	R                    domain.Repost
	RetrievedAtLeastOnce bool
	Id                   int
}

type InMemoryStore struct {
	entries map[int]*entry
	nextId  int
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{make(map[int]*entry, 0), 1}
}

func (s *InMemoryStore) insert(e *entry) {
	e.Id = s.nextId
	s.entries[e.Id] = e
	s.nextId += 1
}
func (s *InMemoryStore) walkAll(handle func(e *entry)) {
	for k, _ := range s.entries {
		handle(s.entries[k])
	}
}
func (s *InMemoryStore) delete(e entry) {
	delete(s.entries, e.Id)
}
