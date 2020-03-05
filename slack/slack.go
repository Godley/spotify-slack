package slack

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/godley/spotify-slack/spotify"
	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
)

type SlackHandler struct {
	Spotify   spotify.Spotify
	skipVoted bool
	skipVotes int
	keepVotes int
	skipTimer *time.Timer
}

func Start(client spotify.Spotify) {

	router := mux.NewRouter()
	router.Handle("/", NewSlackHandler(client))

	fmt.Println("[INFO] Server listening")
	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}

func NewSlackHandler(spotify spotify.Spotify) http.Handler {
	return &SlackHandler{
		Spotify: spotify,
	}
}

func (handler *SlackHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	s, err := slack.SlashCommandParse(req)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !s.ValidateToken(os.Getenv("SLACK_VERIFICATION_TOKEN")) {
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch s.Command {
	case "/spotify":
		response, err := handler.processSpotify(s.Text, s.ChannelID)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		resp.Write([]byte(response))

	default:
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (handler *SlackHandler) processSpotify(text string, channelID string) (string, error) {
	args := strings.Split(text, " ")
	if text == "" {
		return "please specify a command.", nil
	}
	switch args[0] {
	case "add":
		firstTrackFound, err := handler.Spotify.FindTrack(strings.Join(args[1:len(args)-1], " "))
		if err != nil {
			return "", err
		}
		added, err := handler.Spotify.AddToPlaylist(firstTrackFound.ID)
		if err != nil {
			return "", err
		}
		if !added {
			return "Track already in playlist", nil
		}
		return fmt.Sprintf("Added %s to playlist", firstTrackFound.Prompt), nil
	case "playing":
		playingTrack := handler.Spotify.WhatsPlaying()
		return fmt.Sprintf(":cd::musical_note: Now playing %s", playingTrack.Prompt), nil
	case "skip":
		if !handler.skipVoted {
			skipTimer := time.NewTimer(time.Second * 10)
			handler.skipVotes = 1
			handler.skipVoted = true
			go handler.timerExpired(skipTimer.C, channelID)
			return fmt.Sprintf("Voted to skip this track, if no one else votes to keep it in 10 seconds this song will be skipped :arrow_right:"), nil
		} else {
			handler.skipVotes += 1
			return fmt.Sprintf("Voted to skip track. Current votes to skip: %d, votes to keep: %d", handler.skipVotes, handler.keepVotes), nil
		}
	case "keep":
		if !handler.skipVoted {
			return "there is no skip vote in progress", nil
		}
		handler.keepVotes += 1
		return fmt.Sprintf("Voted to keep track. Current votes to skip: %d, votes to keep: %d", handler.skipVotes, handler.keepVotes), nil
	}
	return "", nil
}

func (handler *SlackHandler) timerExpired(channel <-chan time.Time, channelID string) {
	<-channel
	if handler.skipVotes > handler.keepVotes {
		err := handler.Spotify.Skip()
		handler.skipVoted = false
		handler.skipVotes = 0
		handler.keepVotes = 0
		if err != nil {
			fmt.Printf("Failed skipping track: %s\n", err)
		}
	}
}
