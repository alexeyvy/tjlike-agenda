package domain

import "time"

type PublicationId string

type Publication struct {
	Id         PublicationId
	ViewAmount int
	PostedAt   time.Time
}

func NewPublication(id string, viewAmount int, postedAt time.Time) Publication {
	return Publication{Id: PublicationId(id), ViewAmount: viewAmount, PostedAt: postedAt}
}

type Channel struct {
	Id       string
	priority int
}

func NewChannel(id string) Channel {
	return Channel{Id: id}
}

type Repost struct {
	Pub        Publication
	RepostedAt time.Time
	Rate       SuggestionRate
}

func NewRepost(publication Publication, repostedAt time.Time, rate SuggestionRate) Repost {
	return Repost{Pub: publication, RepostedAt: repostedAt, Rate: rate}
}
