package domain

import (
	"errors"
	log "github.com/sirupsen/logrus"
)

type (
	SuggestionRate      float64
	simpleLocalSelector struct{}
)

func (s *simpleLocalSelector) SelectPublication(publications []Publication) (Publication, SuggestionRate) {
	var (
		maxRate                float64
		rate                   float64
		topPublication         Publication
		prevOverweight         = 1.0
		prevPrevOverweight     = 1.0
		prevPrevPrevOverweight = 1.0
	)

	for key, publication := range publications {
		if key <= 2 {
			continue
		}
		prevOverweight = float64(publication.ViewAmount) / float64((publications[key-1]).ViewAmount)
		prevPrevOverweight = float64(publication.ViewAmount) / float64((publications[key-2]).ViewAmount)
		prevPrevPrevOverweight = float64(publication.ViewAmount) / float64((publications[key-3]).ViewAmount)
		if prevOverweight <= 1 {
			continue
		}
		rate = prevOverweight*1.5 + prevPrevOverweight*1.2 + prevPrevPrevOverweight
		if rate > maxRate {
			log.WithFields(log.Fields{
				"id":                 publication.Id,
				"rate":               rate,
				"prevOverweight":     prevOverweight,
				"prevPrevOverweight": prevPrevOverweight,
			}).Debug("New leader of channel")

			maxRate = rate
			topPublication = publication
		}
	}

	if maxRate == 0 {
		return Publication{}, SuggestionRate(0)
	}
	return topPublication, SuggestionRate(maxRate)
}
func NewLocalSelector() *simpleLocalSelector {
	return &simpleLocalSelector{}
}

type LocalSelector interface {
	SelectPublication([]Publication) (Publication, SuggestionRate)
}

type SimpleGlobalSelector struct {
	LocalSelector LocalSelector
}

func NewGlobalSelector(localSelector LocalSelector) *SimpleGlobalSelector {
	return &SimpleGlobalSelector{localSelector}
}

var ErrExhausted = errors.New("all channels exhausted")

const suggestionRateThreshold = 4

func (s *SimpleGlobalSelector) SelectPublication(
	candidates map[Channel][]Publication,
	exists func(publication Publication) bool,
) (Publication, SuggestionRate, error) {
	maxSuggestionRate := 0.0
	var selectedPublication Publication

	// @TODO implement cross-channel selection rules, including chanel priorities

	for _, channelPublications := range candidates {
		msg, suggestionRate := s.LocalSelector.SelectPublication(channelPublications)
		if exists(msg) {
			continue
		}
		if suggestionRate < suggestionRateThreshold {
			continue
		}
		if float64(suggestionRate) > maxSuggestionRate {
			maxSuggestionRate = float64(suggestionRate)
			selectedPublication = msg
		}
	}
	if maxSuggestionRate == 0 {
		return Publication{}, SuggestionRate(0), ErrExhausted
	}

	return selectedPublication, SuggestionRate(maxSuggestionRate), nil
}
