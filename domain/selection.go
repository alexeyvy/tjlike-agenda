package domain

import (
	log "github.com/sirupsen/logrus"
	"sort"
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

	//median := median(viewsDepersonalized)
	if maxRate == 0 {
		return Publication{}, SuggestionRate(0)
	}
	return topPublication, SuggestionRate(maxRate)
}
func NewLocalSelector() *simpleLocalSelector {
	return &simpleLocalSelector{}
}

func median(data []int) int {
	dataCopy := make([]int, len(data))
	copy(dataCopy, data)

	sort.Ints(dataCopy)

	var median int
	l := len(dataCopy)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = (dataCopy[l/2-1] + dataCopy[l/2]) / 2
	} else {
		median = dataCopy[l/2]
	}

	return median
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

type PressureRate int // 0-1
type Requirements struct {
	AlreadySelected []PublicationId
	pressure        PressureRate
}

func (s *SimpleGlobalSelector) SelectPublication(
	candidates map[Channel][]Publication,
	exists func(publication Publication) bool,
) Publication {
	maxSuggestionRate := 0.0
	var selectedPublication Publication
	for _, channelPublications := range candidates {
		msg, suggestionRate := s.LocalSelector.SelectPublication(channelPublications)
		if exists(msg) {
			continue
		}
		if float64(suggestionRate) > maxSuggestionRate {
			maxSuggestionRate = float64(suggestionRate)
			selectedPublication = msg
		}
	}

	return selectedPublication
}
