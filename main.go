package main

import (
	"encoding/json"
	domain "github.com/alexeyvy/tjlike-agenda/domain"
	"github.com/alexeyvy/tjlike-agenda/infra/repost"
	"github.com/alexeyvy/tjlike-agenda/infra/scraping"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Scraper interface {
	ScrapeRecentPublications(domain.Channel) ([]domain.Publication, error)
}

type scraperPoolElement struct {
	platformId string
	s          Scraper
	config     ScraperConfigEntry
}

type ScraperConfigEntry struct {
	Parallel                 bool
	Frequency                int
	PauseBetweenSyncChannels int `yaml:"pause_between_sync_channels"`
	Channels                 []string
}
type preferences struct {
	Scrapers struct {
		Telegram ScraperConfigEntry
	}
}

func initPreferences() preferences {
	log.Info("Reading YAML configuration at config.yaml")

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("YAML config reading error: #%v ", err)
	}
	var preferences preferences
	err = yaml.Unmarshal(yamlFile, &preferences)
	if err != nil {
		log.Fatalf("YAML config unmarshal error: %v", err)
	}
	return preferences
}

type GlobalSelector interface {
	SelectPublication(
		candidates map[domain.Channel][]domain.Publication,
		exists func(publication domain.Publication) bool,
	) (domain.Publication, domain.SuggestionRate, error)
}
type RepostWriterService interface {
	Repost(domain.Publication, domain.SuggestionRate) (domain.Repost, error)
	ExistsForPublication(domain.Publication) bool
}

func main() {
	preferences := initPreferences()

	store := repost.NewFileStore("tjlike_agenda_db.txt")
	runApi(store)

	scraperPool := []scraperPoolElement{
		{
			"telegram",
			&scraping.TelegramScraper{},
			preferences.Scrapers.Telegram,
		},
	}

	var localSelector domain.LocalSelector
	localSelector = domain.NewLocalSelector()
	var selector GlobalSelector
	selector = domain.NewGlobalSelector(localSelector)

	var repostWriter RepostWriterService
	repostWriter = repost.NewService(store)

	for _, scraperPoolEl := range scraperPool {
		scraperPoolEntry := scraperPoolEl
		go func() {
			for {
				collectedPublications := make(map[domain.Channel][]domain.Publication, 1000)
				var (
					scraperWg sync.WaitGroup
					msgMutex  sync.Mutex
				)
				log.Debugf(
					"Preparing to scrape %d channels for platform %s",
					len(scraperPoolEntry.config.Channels),
					scraperPoolEntry.platformId,
				)
				for _, src := range scraperPoolEntry.config.Channels {
					f := func(src string) {
						if scraperPoolEntry.config.Parallel {
							defer scraperWg.Done()
						}
						channel := domain.NewChannel(src)
						publications, err := scraperPoolEntry.s.ScrapeRecentPublications(channel)
						if err != nil {
							log.Errorf("scraping failed on channel %s platform %s: %s", channel.Id, scraperPoolEntry.platformId, err.Error())
							return
						}
						msgMutex.Lock()
						collectedPublications[channel] = publications
						msgMutex.Unlock()
					}
					if scraperPoolEntry.config.Parallel {
						scraperWg.Add(1)
						go f(src)
					} else {
						f(src)
						time.Sleep(time.Second * time.Duration(scraperPoolEntry.config.PauseBetweenSyncChannels))
					}
				}
				scraperWg.Wait()

				log.Debugf("All channels finished for platform %s", scraperPoolEntry.platformId)

				best, suggestionRate, err := selector.SelectPublication(collectedPublications, repostWriter.ExistsForPublication)
				if err == domain.ErrExhausted {
					log.Infof("No trending publications for platform %s so far", scraperPoolEntry.platformId)
				}
				if err == nil {
					if _, err := repostWriter.Repost(best, suggestionRate); err != nil {
						log.Errorf("Repost succeeded, however, there was an error when persisting it in the DB: %s", err)
					}

					log.Infof("Picked most trending publication %s for platform %s", best.Id, scraperPoolEntry.platformId)
				}
				log.Debugf(
					"Sleeping %d seconds before next traversal for platform %s",
					scraperPoolEntry.config.Frequency,
					scraperPoolEntry.platformId,
				)

				time.Sleep(time.Second * time.Duration(scraperPoolEntry.config.Frequency))
			}
		}()
	}
	select {}
}

type RepostReaderService interface {
	PickUpMostTrending(clearUp bool) []domain.Repost
	PurgeIrrelevant() error
}
type RepostApiMessage struct {
	PublicationId string `json:"publication-id"`
	PostedAt      string `json:"posted-at"`
	RepostedAt    string `json:"reposted-at"`
}

func runApi(store repost.Store) {
	var service RepostReaderService
	service = repost.NewService(store)
	r := mux.NewRouter()
	r.HandleFunc("/reposts", func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(200)

		reposts := service.PickUpMostTrending(true)
		repostMessages := make([]RepostApiMessage, len(reposts))
		for k, r := range reposts {
			repostMessages[k] = RepostApiMessage{
				string(r.Pub.Id),
				r.Pub.PostedAt.String(),
				r.RepostedAt.String(),
			}
		}
		jsonOutput, _ := json.Marshal(repostMessages)
		_, err := response.Write(jsonOutput)

		if err != nil {
			log.Error("Error sending API response")
		}
		go func() {
			if err := service.PurgeIrrelevant(); err != nil {
				log.Errorf("cannot purge irrelevant reposts: %s", err.Error())
			}
		}()
	})
	log.Info("Running HTTP API on port 35971")

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(35971), r)
		if err != nil {
			log.Fatalf("failed to run API server")
		}
	}()
}
