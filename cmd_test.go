package bot

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	channel string
	replies []string
	user    *User
)

func responseHandler(target string, message string, sender *User) {
	channel = target
	user = sender
	replies = append(replies, message)
}

func TestMessageReceived(t *testing.T) {
	Convey("Given a new message in the channel", t, func() {
		commands = make(map[string]*customCommand)
		b := New(&Handlers{
			Response: responseHandler,
		})

		Reset(func() {
			channel = ""
			user = &User{Nick: ""}
			replies = []string{}
		})

		Convey("When the command is not registered", func() {
			Convey("It should not post to the channel", func() {
				b.MessageReceived("#go-bot", "!not_a_cmd", &User{})

				So(replies, ShouldBeEmpty)
			})
		})

		Convey("The command can return an error", func() {
			Convey("it sould send the message with the error to the channel", func() {
				cmdError := errors.New("error")
				RegisterCommand("cmd", "", "",
					func(c *Cmd) (string, error) {
						return "", cmdError
					})

				b.MessageReceived("#go-bot", "!cmd", &User{Nick: "user"})

				So(channel, ShouldEqual, "#go-bot")
				So(replies, ShouldResemble,
					[]string{fmt.Sprintf(errorExecutingCommand, "cmd", cmdError.Error())})
			})
		})

		Convey("When the command is valid and registered", func() {
			commands = make(map[string]*customCommand)
			expectedMsg := "msg"
			cmd := "cmd"
			cmdDescription := "Command description"
			cmdExampleArgs := "arg1 arg2"

			RegisterCommand(cmd, cmdDescription, cmdExampleArgs,
				func(c *Cmd) (string, error) {
					return expectedMsg, nil
				})

			Convey("If it is called in the channel, reply on the channel", func() {
				b.MessageReceived("#go-bot", "!cmd", &User{Nick: "user"})

				So(channel, ShouldEqual, "#go-bot")
				So(replies, ShouldResemble, []string{expectedMsg})
			})

			Convey("If it is a private message, reply to the user", func() {
				user = &User{Nick: "go-bot"}
				b.MessageReceived("go-bot", "!cmd", &User{Nick: "sender-nick"})
				So(user.Nick, ShouldEqual, "sender-nick")
			})

			Convey("When the command is help", func() {
				Convey("Display the available commands in the channel", func() {
					b.MessageReceived("#go-bot", "!help", &User{Nick: "user"})

					So(channel, ShouldEqual, "#go-bot")
					So(replies, ShouldResemble, []string{
						fmt.Sprintf(helpAboutCommand, CmdPrefix),
						fmt.Sprintf(availableCommands, "cmd"),
					})
				})

				Convey("If the command exists send a message to the channel", func() {
					b.MessageReceived("#go-bot", "!help cmd", &User{Nick: "user"})

					So(channel, ShouldEqual, "#go-bot")
					So(replies, ShouldResemble, []string{
						fmt.Sprintf(helpDescripton, cmdDescription),
						fmt.Sprintf(helpUsage, CmdPrefix, cmd, cmdExampleArgs),
					})
				})

				Convey("If the command does not exists, display the generic help", func() {
					b.MessageReceived("#go-bot", "!help not_a_command", &User{Nick: "user"})

					So(channel, ShouldEqual, "#go-bot")
					So(replies, ShouldResemble, []string{
						fmt.Sprintf(helpAboutCommand, CmdPrefix),
						fmt.Sprintf(availableCommands, "cmd"),
					})
				})
			})
		})

		Convey("When the command is V2", func() {
			Convey("it should send the message with the error to the channel", func() {
				RegisterCommandV2("cmd", "", "",
					func(c *Cmd) (CmdResult, error) {
						return CmdResult{
							Channel: "#channel",
							Message: "message"}, nil
					})

				b.MessageReceived("#go-bot", "!cmd", &User{Nick: "user"})

				So(channel, ShouldEqual, "#channel")
				So(replies, ShouldResemble, []string{"message"})
			})

			Convey("it should reply to the current channel if the command does not specify one", func() {
				RegisterCommandV2("cmd", "", "",
					func(c *Cmd) (CmdResult, error) {
						return CmdResult{
							Message: "message"}, nil
					})

				b.MessageReceived("#go-bot", "!cmd", &User{Nick: "user"})

				So(channel, ShouldEqual, "#go-bot")
				So(replies, ShouldResemble, []string{"message"})
			})
		})

		Convey("When the command is passive", func() {
			passiveCommands = make(map[string]passiveCmdFunc)

			echo := func(cmd *PassiveCmd) (string, error) {
				return cmd.Raw, nil
			}
			ping := func(cmd *PassiveCmd) (string, error) {
				return "pong", nil
			}
			errored := func(cmd *PassiveCmd) (string, error) {
				return "", errors.New("error")
			}

			RegisterPassiveCommand("echo", echo)
			RegisterPassiveCommand("ping", ping)
			RegisterPassiveCommand("errored", errored)

			Convey("If it is called in the channel, reply on the channel", func() {
				b.MessageReceived("#go-bot", "test", &User{Nick: "user"})

				So(channel, ShouldEqual, "#go-bot")
				So(len(replies), ShouldEqual, 2)
				So(replies, ShouldContain, "test")
				So(replies, ShouldContain, "pong")
			})

			Convey("If it is a private message, reply to the user", func() {
				user = &User{Nick: "go-bot"}
				b.MessageReceived("go-bot", "test", &User{Nick: "sender-nick"})

				So(user.Nick, ShouldEqual, "sender-nick")
				So(len(replies), ShouldEqual, 2)
				So(replies, ShouldContain, "test")
				So(replies, ShouldContain, "pong")
			})
		})
	})
}
