package slack

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/godley/spotify-slack/spotify"

	"github.com/nlopes/slack"
)

const (
	// action is used for slack attament action.
	actionAdd    = "add"
	actionSkip   = "skip"
	actionKeep   = "keep"
	actionCancel = "cancel"
)

type SlackListener struct {
	client  *slack.Client
	spotify spotify.Spotify
	botID   string
}

func InteractiveStart(s spotify.Spotify) {
	// Listening slack event and response
	log.Printf("[INFO] Start slack event listening")
	client := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	slackListener := &SlackListener{
		client:  client,
		spotify: s,
		botID:   "spotify",
	}
	go slackListener.ListenAndResponse()

	// Register handler to receive interactive message
	// responses from slack (kicked by user action)
	// http.Handle("/interaction", interactionHandler{
	// 	verificationToken: os.Getenv("SLACK_VERIFICATION_TOKEN"),
	// })

	log.Printf("[INFO] Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("[ERROR] %s", err)
	}

}

func (s *SlackListener) ListenAndResponse() {
	// Start listening slack events
	rtm := s.client.NewRTM()
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		fmt.Printf("%#v", msg)
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				log.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}

// handleMesageEvent handles message events.
func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent) error {

	// Only response mention to bot. Ignore else.
	if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", s.botID)) {
		return nil
	}

	// Parse message
	m := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
	if len(m) == 0 || m[0] != "add" {
		return fmt.Errorf("invalid message")
	}

	attachment := slack.Attachment{
		Text:       "Which song do you want? :musical_note:",
		Color:      "#f9a41b",
		CallbackID: "add",
		Actions: []slack.AttachmentAction{
			{
				Name:    actionAdd,
				Type:    "select",
				Options: make([]slack.AttachmentActionOption, 0),
			},

			{
				Name:  actionCancel,
				Text:  "Cancel",
				Type:  "button",
				Style: "danger",
			},
		},
	}
	options, err := s.spotify.FindTrack(m[1], strings.Join(m[2:len(m)-1], " "))
	if err != nil {
		return err
	}
	if len(options) == 1 {
		added, err := s.spotify.AddToPlaylist(options[0].ID)
		if err != nil {
			return err
		}
		a := slack.Attachment{
			Pretext: "some pretext",
			Text:    "some text",
		}
		if !added {
			s.client.PostMessage(ev.Channel, slack.MsgOptionText("Track already in playlist", false), slack.MsgOptionAttachments(a))
		}

		s.client.PostMessage(ev.Channel, slack.MsgOptionText("Track added to playlist", false), slack.MsgOptionAttachments(a))
	} else {
		for _, option := range options {
			attachment.Actions[0].Options = append(attachment.Actions[0].Options, slack.AttachmentActionOption{
				Text:  option.Prompt,
				Value: string(option.ID),
			})
		}
	}

	// value is passed to message handler when request is approved.

	if _, _, err := s.client.PostMessage(ev.Channel, slack.MsgOptionText("", false), slack.MsgOptionAttachments(attachment)); err != nil {
		return fmt.Errorf("failed to post message: %s", err)
	}

	return nil
}
