# animelist-scrobbler
A Plex webhook MyAnimeList scrobbler written in Go

## Installation
`go get github.com/djcrock/animelist-scrobbler`

## Arguments

| Argument     | Required | Description                                                                                                                   |
|--------------|----------|-------------------------------------------------------------------------------------------------------------------------------|
| `-maluser`   | **Yes**  | MyAnimeList username                                                                                                          |
| `-malpass`   | **Yes**  | MyAnimeList password                                                                                                          |
| `-plexurl`   | **Yes**  | URL of the Plex server                                                                                                        |
| `-plextoken` | **Yes**  | Plex [authentication token](https://support.plex.tv/hc/en-us/articles/204059436-Finding-an-authentication-token-X-Plex-Token) |
| `-plexuser`  |          | Username of the Plex user for whom to scrobble. Will scrobble activity of all users if omitted.                               |
| `-port`      |          | Port on which to run the service. Defaults to `8080`.                                                                         |
| `-test`      |          | Trigger scrobbling on pause as well, but do not update MyAnimeList                                                            |

## Usage
This scrobbler relies on Plex webhooks, so a Plex Pass subscription is required.

### Start Service
`animelist-scrobbler -maluser $MAL_USER -malpass $MAL_PASS -plexurl $PLEX_URL -plextoken $PLEX_TOKEN -plexuser $PLEX_USER`

### Configure Webhooks
Configure Plex to send [webhooks](https://support.plex.tv/hc/en-us/articles/115002267687-Webhooks) to the scrobbler.

### Add Metadata to Series
In order for the scrobbler to properly identify series and episode numbers, update the "Summary" of each series of each anime you want to scrobble in the following format:
```json
{"myAnimeListID": 34494, "firstEpisode": 1}
```
where `myAnimeListID` is the ID of the anime on [myanimelist.net](https://myanimelist.net), and `firstEpisode` is the MyAnimeList episode number of the "first" episode of the season in Plex (this helps match the correct episode when MyAnimeList's seasons and Plex's seasons don't line up).
