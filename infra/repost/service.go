package repost

import (
	"github.com/alexeyvy/tjlike-agenda/domain"
	"time"
)

type Store interface {
	insert(*entry)
	walkAll(func(e *entry))
	delete(entry)
}
type LockableStore interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}
type PersistentStore interface {
	Pull() error
	Push() error
}

type service struct {
	s Store
}

func NewService(store Store) *service {
	persistentStore, isPersistentStore := store.(PersistentStore)
	if isPersistentStore {
		if err := persistentStore.Pull(); err != nil {
			if err != ErrDBNotInited {
				panic("cannot create repost service due to failed persistentStore.Pull(): " + err.Error())
			}
		}
	}

	return &service{store}
}

var now = time.Now

const hoursBeforePurge = 48

type PersistDBFailed struct {
	originalError error
}

func (e *PersistDBFailed) Error() string {
	return "cannot persist DB: " + e.originalError.Error()
}

func (r *service) Repost(publication domain.Publication, rate domain.SuggestionRate) (domain.Repost, error) {
	repost := domain.NewRepost(publication, now(), rate)
	dbEntry := &entry{repost, false, 0}

	lockableStore, isLockable := r.s.(LockableStore)
	if isLockable {
		lockableStore.Lock()
		defer lockableStore.Unlock()
	}
	r.s.insert(dbEntry)

	persistentStore, isPersistentStore := r.s.(PersistentStore)
	if isPersistentStore {
		if err := persistentStore.Push(); err != nil {
			return repost, &PersistDBFailed{err}
		}
	}

	return repost, nil
}

func (r *service) PickUpMostTrending(clearUp bool) []domain.Repost {
	reposts := make([]domain.Repost, 0)

	lockableStore, isLockable := r.s.(LockableStore)
	if isLockable {
		lockableStore.RLock()
		defer lockableStore.RUnlock()
	}

	r.s.walkAll(func(e *entry) {
		if e.RetrievedAtLeastOnce == false {
			reposts = append(reposts, e.R)
			if clearUp {
				e.RetrievedAtLeastOnce = true
			}
		}
	})

	return reposts
}

type PurgeUnreadError struct{}

func (e PurgeUnreadError) Error() string {
	return "storage can't be emptied, as there are reposts that have not yet been pulled"
}

func (r *service) PurgeIrrelevant() error {
	repostedAtThreshold := now().Add(-1 * hoursBeforePurge * time.Hour)
	entriesForDeletion := make([]entry, 0)
	var unreadFound bool

	lockableStore, isLockable := r.s.(LockableStore)
	if isLockable {
		lockableStore.Lock()
		defer lockableStore.Unlock()
	}

	r.s.walkAll(func(e *entry) {
		if e.R.RepostedAt.Before(repostedAtThreshold) {
			if e.RetrievedAtLeastOnce == false {
				unreadFound = true
			}
			entriesForDeletion = append(entriesForDeletion, *e)
		}
	})

	if unreadFound {
		return &PurgeUnreadError{}
	}

	for _, e := range entriesForDeletion {
		r.s.delete(e)
	}

	persistentStore, isPersistentStore := r.s.(PersistentStore)
	if isPersistentStore {
		if err := persistentStore.Push(); err != nil {
			return &PersistDBFailed{err}
		}
	}

	return nil
}

func (r *service) ExistsForPublication(publication domain.Publication) bool {
	var exists bool

	lockableStore, isLockable := r.s.(LockableStore)
	if isLockable {
		lockableStore.RLock()
		defer lockableStore.RUnlock()
	}

	r.s.walkAll(func(e *entry) {
		if e.R.Pub.Id == publication.Id {
			exists = true
		}
	})

	return exists
}
