package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jrudio/go-plex-client"
	"github.com/nstratos/go-myanimelist/mal"
)

// CustomPlexMetadata is included in a series' "Summary" field
type CustomPlexMetadata struct {
	MyAnimeListID int `json:"myAnimeListID"`
	FirstEpisode  int `json:"firstEpisode"`
}

var (
	malClient  *mal.Client
	plexClient *plex.Plex
	plexUser   string
	malUser    string
	port       int
	testMode   bool
)

func main() {
	var malUserName, malPass, plexURL, plexToken string
	flag.StringVar(&malUserName, "maluser", "", "MyAnimelist username")
	flag.StringVar(&malPass, "malpass", "", "MyAnimelist password")
	flag.StringVar(&plexURL, "plexurl", "", "URL of the Plex server")
	flag.StringVar(&plexToken, "plextoken", "", "Plex authentication token")
	flag.StringVar(&plexUser, "plexuser", "", "Username of the Plex user for whom to scrobble. Will scrobble activity of all users if omitted.")
	flag.IntVar(&port, "port", 8080, "Port on which to run the service. Defaults to 8080.")
	flag.BoolVar(&testMode, "test", false, "Trigger scrobbling on pause as well, but do not update MyAnimeList")
	flag.Parse()

	if plexURL == "" || plexToken == "" {
		log.Fatal("Plex URL and token are required")
	}

	if malUserName == "" || malPass == "" {
		log.Fatal("MyAnimeList username and password are required")
	}

	log.Print("Connecting to Plex server...")
	pc, err := plex.New(plexURL, plexToken)
	if err != nil {
		log.Fatalf("Could not initialize Plex API: %v", err)
	}
	plexClient = pc
	_, err = plexClient.Test()
	if err != nil {
		log.Fatalf("Could not connect to Plex server: %v", err)
	}
	log.Print("Connected to Plex server")

	log.Print("Verifying MyAnimeList account...")
	malClient = mal.NewClient(nil)
	malClient.SetCredentials(malUserName, malPass)
	user, _, err := malClient.Account.Verify()
	if err != nil {
		log.Fatalf("Failed to log in to MyAnimeList: %s", err)
	}
	log.Printf("Verified MyAnimeList user %v", user.Username)
	malUser = user.Username

	mux := http.NewServeMux()
	wh := plex.NewWebhook()
	wh.OnScrobble(handleScrobbleWebhook)
	if testMode {
		wh.OnPause(handleScrobbleWebhook)
	}
	mux.HandleFunc("/plex", wh.Handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), mux))
}

func handleScrobbleWebhook(w plex.Webhook) {
	if plexUser != "" && w.Account.Title != plexUser {
		log.Printf("Hook received for a user other than %s. Ignoring.", plexUser)
		return
	}

	log.Printf(
		"User %s on %s finished watching %s - %s - %s. Scrobbling!",
		w.Account.Title,
		w.Server.Title,
		w.Metadata.GrandparentTitle,
		w.Metadata.ParentTitle,
		w.Metadata.Title,
	)
	scrobble(w)
}

func scrobble(w plex.Webhook) {
	customMetadata, err := getCustomPlexMetadata(w.Metadata.ParentRatingKey)
	if err != nil {
		log.Printf("Failed to retrieve custom Plex metadata: %v", err)
		return
	}
	episode := w.Metadata.Index + (customMetadata.FirstEpisode - 1)

	list, _, err := malClient.Anime.List(malUser)
	if err != nil {
		log.Printf("Unable to retrieve user's anime list: %v", err)
		return
	}

	var anime mal.Anime
	for _, a := range list.Anime {
		if a.SeriesAnimeDBID == customMetadata.MyAnimeListID {
			anime = a
		}
	}
	if anime.SeriesAnimeDBID != customMetadata.MyAnimeListID {
		log.Printf("No anime found for ID %d", customMetadata.MyAnimeListID)
		return
	}
	log.Printf("Found anime: %s (%d) episode %d", anime.SeriesTitle, customMetadata.MyAnimeListID, episode)

	rewatching, err := strconv.Atoi(anime.MyRewatching)
	if err != nil {
		log.Printf("Could not parse 'rewatching' status: %v", err)
		return
	}

	if anime.MyStatus == mal.StatusCompleted && rewatching == 0 {
		log.Printf("Starting re-watch")
		rewatching = 1
	} else if anime.MyWatchedEpisodes >= episode {
		log.Printf("Completed episode %d, but MAL entry has %d episodes watched - doing nothing", episode, anime.MyWatchedEpisodes)
		return
	}

	var status int
	if anime.MyStatus == mal.StatusCompleted || episode == anime.SeriesEpisodes {
		status = mal.StatusCompleted
	} else {
		status = mal.StatusWatching
	}

	ae := mal.AnimeEntry{
		Episode:          episode,
		Status:           strconv.Itoa(status),
		Score:            anime.MyScore,
		EnableRewatching: rewatching,
		Tags:             anime.MyTags,
	}

	if anime.MyStartDate == "0000-00-00" {
		ae.DateStart = time.Now().Format("01022006")
	}
	if status == mal.StatusCompleted && anime.MyFinishDate == "0000-00-00" {
		ae.DateFinish = time.Now().Format("01022006")
	}

	if testMode {
		log.Print("Test mode enabled, not updating MyAnimeList")
		return
	}
	_, err = malClient.Anime.Update(customMetadata.MyAnimeListID, ae)
	if err != nil {
		log.Printf("Failed to update MyAnimeList entry: %v", err)
	}
	log.Printf("Updated MyAnimeList entry: %s episode %d watched", anime.SeriesTitle, episode)
}

// getCustomPlexMetadata retrieves metadata stored as JSON in the "Summary" field for the season
func getCustomPlexMetadata(key string) (*CustomPlexMetadata, error) {
	m, err := plexClient.GetMetadata(key)
	if err != nil {
		return nil, fmt.Errorf("api error: %v", err)
	}
	customMetadata := &CustomPlexMetadata{}
	err = json.Unmarshal([]byte(m.Directory.Summary), customMetadata)
	if err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}
	return customMetadata, nil
}
