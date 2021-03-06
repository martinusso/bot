// Package bot provides a simple to use IRC bot
package bot

import (
	"log"
	"math/rand"
	"time"

	"github.com/robfig/cron"
)

const (
	// CmdPrefix is the prefix used to identify a command.
	// !hello whould be identified as a command
	CmdPrefix = "!"
)

// Bot handles the bot instance
type Bot struct {
	handlers *Handlers
	cron     *cron.Cron
}

// ResponseHandler must be implemented by the protocol to handle the bot responses
type ResponseHandler func(target, message string, sender *User)

// Handlers that must be registered to receive callbacks from the bot
type Handlers struct {
	Response ResponseHandler
}

// New configures a new bot instance
func New(h *Handlers) *Bot {
	b := &Bot{
		handlers: h,
		cron:     cron.New(),
	}
	b.startPeriodicCommands()
	return b
}

func (b *Bot) startPeriodicCommands() {
	for _, config := range periodicCommands {
		b.cron.AddFunc(config.CronSpec, func() {
			message, err := config.CmdFunc()
			if err != nil {
				log.Print("Periodic command failed ", err)
				return
			}
			if message != "" {
				for _, channel := range config.Channels {
					b.handlers.Response(channel, message, nil)
				}
			}
		})
	}
	if len(b.cron.Entries()) == 1 {
		b.cron.Start()
	}
}

// MessageReceived must be called by the protocol upon receiving a message
func (b *Bot) MessageReceived(channel string, text string, sender *User) {
	command := parse(text, channel, sender)
	if command == nil {
		b.executePassiveCommands(&PassiveCmd{
			Raw:     text,
			Channel: channel,
			User:    sender,
		})
		return
	}

	switch command.Command {
	case helpCommand:
		b.help(command)
	default:
		b.handleCmd(command)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
