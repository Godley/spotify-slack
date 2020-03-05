package spotify

import (
	"fmt"
	"strings"

	"github.com/zmb3/spotify"
)

type Spotify interface {
	AddToPlaylist(trackID spotify.ID) (bool, error)
	FindTrack(query string) (Result, error)
	WhatsPlaying() Result
	Skip() error
}

type SpotifyClient struct {
	spotify  *spotify.Client
	Playlist *spotify.FullPlaylist
}

func NewSpotifyClient(client *spotify.Client, playlistID string) (Spotify, error) {
	spotClient := &SpotifyClient{
		spotify: client,
	}
	playlist, err := client.GetPlaylist(spotify.ID(playlistID))
	if err != nil {
		return nil, err
	}
	spotClient.Playlist = playlist

	return spotClient, nil
}

type Result struct {
	ID     spotify.ID
	Prompt string
}

var defaultTrackId spotify.ID = "4uLU6hMCjMI75M1A2tKUQC"

func (s *SpotifyClient) WhatsPlaying() Result {
	playing, err := s.spotify.PlayerCurrentlyPlaying()
	if err != nil {
		return Result{
			ID:     defaultTrackId,
			Prompt: "I don't know mate",
		}
	}

	artistNames := []string{}
	for _, artist := range playing.Item.Artists {
		artistNames = append(artistNames, artist.Name)
	}
	return Result{
		ID:     playing.Item.ID,
		Prompt: fmt.Sprintf("%s by %s", playing.Item.Name, strings.Join(artistNames, ",")),
	}
}

func (s *SpotifyClient) FindTrack(query string) (Result, error) {
	results, err := s.spotify.Search(query, spotify.SearchTypeTrack)
	if err != nil {
		return Result{}, err
	}
	if results.Tracks != nil {
		artistNames := make([]string, 0)
		for _, artist := range results.Tracks.Tracks[0].Artists {
			artistNames = append(artistNames, artist.Name)
		}
		return Result{
			ID:     results.Tracks.Tracks[0].ID,
			Prompt: fmt.Sprintf("%s by %s", results.Tracks.Tracks[0].Name, strings.Join(artistNames, ",")),
		}, nil
	}
	return Result{}, fmt.Errorf("No results")
}

func (s *SpotifyClient) AddToPlaylist(trackID spotify.ID) (bool, error) {
	trackInPlaylist, err := s.isTrackInPlaylist(trackID)
	if err != nil {
		return false, err
	}
	if trackInPlaylist {
		return false, nil
	}
	_, err = s.spotify.AddTracksToPlaylist(s.Playlist.ID, trackID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SpotifyClient) isTrackInPlaylist(trackID spotify.ID) (bool, error) {
	inPage := false
	tracks, err := s.spotify.GetPlaylistTracks(s.Playlist.ID)
	if err != nil {
		return true, err
	}
	for true {
		inPage = isTrackInPage(trackID, *tracks)
		if inPage {
			return false, nil
		}
		err := s.spotify.NextPage(tracks)
		if err != nil && err == spotify.ErrNoMorePages {
			return false, nil
		} else if err != nil {
			// TODO log stuff here
			return false, err
		}
	}
	return false, nil
}

func isTrackInPage(trackID spotify.ID, page spotify.PlaylistTrackPage) bool {
	for _, t := range page.Tracks {
		if t.Track.ID == trackID {
			return true
		}
	}
	return false
}

func (s *SpotifyClient) Skip() error {
	return s.spotify.Next()
}
