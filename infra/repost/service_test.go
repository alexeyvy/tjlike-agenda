package repost

import (
	"github.com/alexeyvy/tjlike-agenda/domain"
	"testing"
	"time"
)

func TestSaveAndPickup(t *testing.T) {
	t.Parallel()

	s := NewService(NewInMemoryStore())

	postedAt, _ := time.Parse(time.RFC822, "02 Jan 06 15:04 MST")
	s.Repost(domain.NewPublication("platform/id", 10300, postedAt))

	reposts := s.PickUpMostTrending(false)
	if len(reposts) != 1 {
		t.Errorf(
			"Picking up saved repost failed",
		)
	}
	reposts = s.PickUpMostTrending(true)
	if len(reposts) != 1 {
		t.Errorf(
			"Dry mode failed, cannot pick up repetatively",
		)
	}
	reposts = s.PickUpMostTrending(false)
	if len(reposts) != 0 {
		t.Errorf(
			"%d entries left off after picked up in wet mode", len(reposts),
		)
	}
}

func TestPurgeIrrelevant(t *testing.T) {
	t.Parallel()

	var currentTime string
	now = func() time.Time { a, _ := time.Parse(time.RFC822, currentTime); return a }

	s := NewService(NewInMemoryStore())

	currentTime = "02 Jan 06 15:06 MST"
	postedAt, _ := time.Parse(time.RFC822, "02 Jan 06 15:04 MST")
	s.Repost(domain.NewPublication("platform/id", 10300, postedAt))

	currentTime = "06 Jan 06 15:06 MST"
	err := s.PurgeIrrelevant()
	if err == nil {
		t.Errorf(
			"PurgeUnreadError expected",
		)
	}
}
