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

type service struct {
	s Store
}

func NewService(store Store) *service {
	return &service{store}
}

var now = time.Now

func (r *service) Repost(publication domain.Publication) domain.Repost {
	repost := domain.NewRepost(publication, now())
	dbEntry := &entry{repost, false, 0}
	r.s.insert(dbEntry)

	return repost
}

func (r *service) PickUpMostTrending(clearUp bool) []domain.Repost {
	reposts := make([]domain.Repost, 0)
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
	return "Storage can't be emptied, as there are reposts that have not yet been pulled."
}

func (r *service) PurgeIrrelevant() error {
	repostedAtThreshold := now().Add(-48 * time.Hour)
	entriesForDeletion := make([]entry, 0)
	var unreadFound bool
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

	return nil
}

func (r *service) ExistsForPublication(publication domain.Publication) bool {
	var exists bool

	r.s.walkAll(func(e *entry) {
		if e.R.Pub.Id == publication.Id {
			exists = true
		}
	})

	return exists
}
