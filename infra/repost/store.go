package repost

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexeyvy/tjlike-agenda/domain"
	"os"
	"sync"
)

type entry struct {
	R                    domain.Repost
	RetrievedAtLeastOnce bool
	Id                   int
}

type InMemoryStore struct {
	entries map[int]*entry
	nextId  int
	sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{make(map[int]*entry, 0), 1, sync.RWMutex{}}
}

func (s *InMemoryStore) insert(e *entry) {
	e.Id = s.nextId
	s.entries[e.Id] = e
	s.nextId += 1
}
func (s *InMemoryStore) walkAll(handle func(e *entry)) {
	for k := range s.entries {
		handle(s.entries[k])
	}
}
func (s *InMemoryStore) delete(e entry) {
	delete(s.entries, e.Id)
}

type FileStore struct {
	InMemoryStore
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{*NewInMemoryStore(), path}
}

type jsonRepresentation struct {
	Entries map[int]*entry
	NextId  int
}

var ErrDBNotInited = errors.New("cannot initialize store as the source does not exist")

func (fs *FileStore) Pull() error {
	if _, err := os.Stat(fs.path); errors.Is(err, os.ErrNotExist) {
		return ErrDBNotInited
	}
	data, err := os.ReadFile(fs.path)
	var jr jsonRepresentation
	err = json.Unmarshal(data, &jr)
	if err != nil {
		return fmt.Errorf("cannot unmarshal store: %w", err)
	}
	fs.entries = jr.Entries
	fs.nextId = jr.NextId
	return nil
}
func (fs *FileStore) Push() error {
	jr := jsonRepresentation{fs.entries, fs.nextId}
	jsonEncoded, _ := json.Marshal(jr)
	err := os.WriteFile(fs.path, jsonEncoded, 0644)
	if err != nil {
		return fmt.Errorf("cannot write DB file: %w", err)
	}
	return nil
}
