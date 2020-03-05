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
	Spotify     spotify.Spotify
	SlackWriter MessageWriter
	channelID   string
	skipVoted   bool
	skipVotes   int
	keepVotes   int
	skipTimer   *time.Timer
}

func Start(client spotify.Spotify, token, channelID string) {

	router := mux.NewRouter()
	router.Handle("/", NewSlackHandler(client, token, channelID))

	fmt.Println("[INFO] Server listening")
	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}

func NewSlackHandler(spotify spotify.Spotify, token, channelID string) http.Handler {
	handler := &SlackHandler{
		Spotify:     spotify,
		SlackWriter: NewPoster(token, channelID),
		channelID:   channelID,
	}
	go handler.SlackWriter.StartPoster()
	return handler
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
	if s.ChannelID != handler.channelID {
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
			return fmt.Sprintf(":guitar: %s already in playlist", firstTrackFound.Prompt), nil
		}
		handler.SlackWriter.Write(fmt.Sprintf(":musical_keyboard: Added %s to playlist", firstTrackFound.Prompt))
		return "", nil
	case "playing":
		playingTrack := handler.Spotify.WhatsPlaying()
		if playingTrack.ID == "" {
			handler.SlackWriter.Write(":upside_down_face: nothing is currently playing in the office")
		} else {
			handler.SlackWriter.Write(fmt.Sprintf(":cd::musical_note: _Now playing:_ %s", playingTrack.Prompt))
		}
		return "", nil
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
