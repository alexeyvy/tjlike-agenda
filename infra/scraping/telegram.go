package scraping

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/alexeyvy/tjlike-agenda/domain"
	"net/http"
	"strconv"
	"time"
)

type TelegramScraper struct {
}

const maxMsgNum = 20

type telegramPublication struct {
	id         string
	viewAmount string
	postedAt   string
}

func dehumanizeViewNumber(humanized string) int {
	var viewAmount int

	switch humanized[len(humanized)-1:] {
	case "K":
		f, _ := strconv.ParseFloat(humanized[:len(humanized)-1], 64)
		viewAmount = int(f * 1000)
	case "M":
		f, _ := strconv.ParseFloat(humanized[:len(humanized)-1], 64)
		viewAmount = int(f * 1000000)
	default:
		viewAmount, _ = strconv.Atoi(humanized)
	}
	return viewAmount
}

func (tm telegramPublication) generalize() domain.Publication {
	postedAt, _ := time.Parse(time.RFC3339, tm.postedAt)
	return domain.NewPublication(tm.id, dehumanizeViewNumber(tm.viewAmount), postedAt.UTC())
}

func (s *TelegramScraper) ScrapeRecentPublications(channel domain.Channel) ([]domain.Publication, error) {
	unformattedPublications, err := s.scrapeRecentPublications(channel.Id)
	if err != nil {
		return nil, err
	}

	formattedPublications := make([]domain.Publication, 0, maxMsgNum)
	for publicationKey := range unformattedPublications {
		formattedPublications = append(formattedPublications, unformattedPublications[publicationKey].generalize())
	}
	return formattedPublications, nil
}

func (s *TelegramScraper) scrapeRecentPublications(channelId string) ([]telegramPublication, error) {
	url := fmt.Sprintf("https://t.me/s/%s", channelId)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New("HTTP request failed")
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot parse HTTP response: %w", err)
	}

	var publications []telegramPublication

	container := doc.Find("div.tgme_widget_message_wrap")
	if container.Length() == 0 {
		return nil, errors.New("no publications found in div.tgme_widget_message_wrap. this may be private or non-existent channel")
	}

	var e error
	container.Each(func(i int, s *goquery.Selection) {
		idContainer := s.Find("div.tgme_widget_message")
		var id string
		if idContainer.Length() == 1 {
			var exists bool
			id, exists = idContainer.Attr("data-post")
			if !exists {
				e = fmt.Errorf("one of publications has no ID")
				return
			}
		} else {
			e = fmt.Errorf("container ID not found")
			return
		}

		viewsContainer := s.Find(".tgme_widget_message_views")
		// this is an expected issue, skip one without views
		if viewsContainer.Length() == 0 {
			return
		}
		viewAmount := viewsContainer.Text()

		postedAtContainer := s.Find(".tgme_widget_message_date time")
		var postedAt string
		if postedAtContainer.Length() == 1 {
			var exists bool
			postedAt, exists = postedAtContainer.Attr("datetime")
			if !exists {
				e = fmt.Errorf("one of publications has no POSTED AT")
				return
			}
		} else {
			e = fmt.Errorf("container POSTED AT not found")
			return
		}

		publications = append(publications, telegramPublication{id: id, viewAmount: viewAmount, postedAt: postedAt})
	})
	if e != nil {
		return publications, e
	}

	return publications, nil
}
