package slack

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/godley/spotify-slack/spotify"
	"github.com/gorilla/mux"
	"github.com/nlopes/slack"
)

type SlackHandler struct {
	Spotify spotify.Spotify
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

	if !s.ValidateToken(os.Getenv("SLACK_TOKEN")) {
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch s.Command {
	case "/spotify":
		response, err := handler.processSpotify(s.Text)
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

func (handler *SlackHandler) processSpotify(text string) (string, error) {
	args := strings.Split(text, " ")
	if len(args) < 2 {
		return "please specify a command.", nil
	}
	switch args[0] {
	case "add":
		options, err := handler.Spotify.FindTrack(args[1], strings.Join(args[2:len(args)-1], " "))
		if err != nil {
			return "", err
		}
		if len(options) == 1 {
			added, err := handler.Spotify.AddToPlaylist(options[0].ID)
			if err != nil {
				return "", err
			}
			if !added {
				return "Track already in playlist", nil
			}
			return "Added to playlist", nil
		} else {
			response := "Multiple tracks found with title: %s and artist: %s. Select one:\n%s"
			optionsText := ""
			for _, option := range options {
				optionsText += fmt.Sprintf("%s\n", option.Prompt)
			}
			return fmt.Sprintf(response, args[1], args[2], optionsText), nil
		}
	}
	return "", nil
}
