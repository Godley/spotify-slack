package main

import (
	"github.com/godley/spotify-slack/spotify"
)

func main() {
	_, _ = spotify.NewSpotifyClient(spotify.SpotifyClientOpts{})
}
