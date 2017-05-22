FROM golang

ADD . /go/src/github.com/djcrock/animelist-scrobbler

RUN go install github.com/djcrock/animelist-scrobbler

ENTRYPOINT /go/bin/animelist-scrobbler -maluser $MAL_USER -malpass $MAL_PASS -plexurl $PLEX_URL -plextoken $PLEX_TOKEN -plexuser $PLEX_USER
