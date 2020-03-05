package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

type MessageWriter interface {
	StartPoster()
	Write(message string)
}
type Poster struct {
	messages  chan string
	channelID string
	api       *slack.Client
}

func NewPoster(token, channelID string) MessageWriter {
	return &Poster{
		messages:  make(chan string, 10),
		channelID: channelID,
		api:       slack.New(token),
	}
}
func (p *Poster) Write(message string) {
	p.messages <- message
}

func (p *Poster) StartPoster() {
	for {
		msg := <-p.messages
		attachment := slack.Attachment{}

		_, _, err := p.api.PostMessage(p.channelID, slack.MsgOptionText(msg, false), slack.MsgOptionAttachments(attachment))
		if err != nil {
			fmt.Printf("Failed posting msg: %s", err)
		}
	}
}
