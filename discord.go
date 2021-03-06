package whisperer

import (
	"fmt"
	"io"
	"log"
	"regexp"

	"github.com/jonas747/discordgo"
)

// DiscordServiceName is the service name for the Discord service.
const DiscordServiceName string = "Discord"

// DiscordMessage is a Message wrapper around discordgo.Message.
type DiscordMessage struct {
	DiscordgoMessage *discordgo.Message
	MessageType      MessageType
}

// Channel returns the channel id for this message.
func (m *DiscordMessage) Channel() string {
	return m.DiscordgoMessage.ChannelID
}

// UserName returns the user name for this message.
func (m *DiscordMessage) UserName() string {
	if m.DiscordgoMessage.Author == nil {
		return ""
	}
	return m.DiscordgoMessage.Author.Username
}

// UserID returns the user id for this message.
func (m *DiscordMessage) UserID() string {
	if m.DiscordgoMessage.Author == nil {
		return ""
	}
	return m.DiscordgoMessage.Author.ID
}

// Message returns the message content for this message.
func (m *DiscordMessage) Message() string {
	return m.DiscordgoMessage.ContentWithMentionsReplaced()
}

// RawMessage returns the raw message content for this message.
func (m *DiscordMessage) RawMessage() string {
	return m.DiscordgoMessage.Content
}

// MessageID returns the message ID for this message.
func (m *DiscordMessage) MessageID() string {
	return m.DiscordgoMessage.ID
}

// Type returns the type of message.
func (m *DiscordMessage) Type() MessageType {
	return m.MessageType
}

// Discord is a Service provider for Discord.
type Discord struct {
	args        []interface{}
	messageChan chan Message

	Session             *discordgo.Session
	OwnerUserID         string
	ApplicationClientID string
}

// NewDiscord creates a new discord service.
func NewDiscord(args ...interface{}) *Discord {
	return &Discord{
		args:        args,
		messageChan: make(chan Message, 200),
	}
}

var channelIDRegex = regexp.MustCompile("<#[0-9]*>")

func (d *Discord) replaceChannelNames(message *discordgo.Message) {
	message.Content = channelIDRegex.ReplaceAllStringFunc(message.Content, func(str string) string {
		c, err := d.Session.State.Channel(str[2 : len(str)-1])
		if err != nil {
			return str
		}

		return "#" + c.Name
	})
}

var roleIDRegex = regexp.MustCompile("<@&[0-9]*>")

func (d *Discord) replaceRoleNames(message *discordgo.Message) {
	message.Content = roleIDRegex.ReplaceAllStringFunc(message.Content, func(str string) string {
		roleID := str[3 : len(str)-1]

		c, err := d.Session.State.Channel(message.ChannelID)
		if err != nil {
			return str
		}

		g, err := d.Session.State.Guild(c.GuildID)
		if err != nil {
			return str
		}

		for _, r := range g.Roles {
			if r.ID == roleID {
				return "@" + r.Name
			}
		}

		return str
	})
}

func (d *Discord) onMessageCreate(s *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Content == "" {
		return
	}

	d.replaceChannelNames(message.Message)
	d.replaceRoleNames(message.Message)

	d.messageChan <- &DiscordMessage{message.Message, MessageTypeCreate}
}

func (d *Discord) onMessageUpdate(s *discordgo.Session, message *discordgo.MessageUpdate) {
	if message.Content == "" {
		return
	}

	d.replaceChannelNames(message.Message)

	d.messageChan <- &DiscordMessage{message.Message, MessageTypeUpdate}
}

func (d *Discord) onMessageDelete(s *discordgo.Session, message *discordgo.MessageDelete) {
	d.messageChan <- &DiscordMessage{message.Message, MessageTypeDelete}
}

// Name returns the name of the service.
func (d *Discord) Name() string {
	return DiscordServiceName
}

// Open opens the service and returns a channel which all messages will be sent on.
func (d *Discord) Open() (<-chan Message, error) {
	var err error

	d.Session, err = discordgo.New(d.args...)
	if err != nil {
		return nil, err
	}

	d.Session.AddHandler(d.onMessageCreate)
	d.Session.AddHandler(d.onMessageUpdate)
	d.Session.AddHandler(d.onMessageDelete)

	d.Session.Open()

	return d.messageChan, nil
}

// IsMe returns whether or not a message was sent by the bot.
func (d *Discord) IsMe(message Message) bool {
	if d.Session.State.User == nil {
		return false
	}
	return message.UserID() == d.Session.State.User.ID
}

// SendMessage sends a message.
func (d *Discord) SendMessage(channel, message string) error {
	if _, err := d.Session.ChannelMessageSend(channel, message); err != nil {
		log.Println("Error sending discord message: ", err)
		return err
	}
	return nil
}

// DeleteMessage deletes a message.
func (d *Discord) DeleteMessage(channel, messageID string) error {
	return d.Session.ChannelMessageDelete(channel, messageID)
}

// SendFile sends a file.
func (d *Discord) SendFile(channel, name string, r io.Reader) error {
	if _, err := d.Session.ChannelFileSend(channel, name, r); err != nil {
		log.Println("Error sending discord message: ", err)
		return err
	}
	return nil
}

// BanUser bans a user.
func (d *Discord) BanUser(channel, userID string, duration int) error {
	return d.Session.GuildBanCreate(channel, userID, 0)
}

// UnbanUser unbans a user.
func (d *Discord) UnbanUser(channel, userID string) error {
	return d.Session.GuildBanDelete(channel, userID)
}

// UserName returns the bots name.
func (d *Discord) UserName() string {
	if d.Session.State.User == nil {
		return ""
	}
	return d.Session.State.User.Username
}

// Join accept an invite or return an error.
// If AlreadyJoinedError is return, @me has already accepted that invite.
func (d *Discord) Join(join string) error {
	if i, err := d.Session.Invite(join); err == nil {
		if _, err := d.Session.State.Guild(i.Guild.ID); err == nil {
			return ErrAlreadyJoined
		}
	}

	if _, err := d.Session.InviteAccept(join); err != nil {
		return err
	}
	return nil
}

// Typing sets that the bot is typing.
func (d *Discord) Typing(channel string) error {
	return d.Session.ChannelTyping(channel)
}

// PrivateMessage will send a private message to a user.
func (d *Discord) PrivateMessage(userID, message string) error {
	c, err := d.Session.UserChannelCreate(userID)
	if err != nil {
		return err
	}
	return d.SendMessage(c.ID, message)
}

// SupportsPrivateMessages returns whether the service supports private messages.
func (d *Discord) SupportsPrivateMessages() bool {
	return true
}

// SupportsMultiline returns whether the service supports multiline messages.
func (d *Discord) SupportsMultiline() bool {
	return true
}

// CommandPrefix returns the command prefix for the service.
func (d *Discord) CommandPrefix() string {
	return fmt.Sprintf("@%s ", d.UserName())
}

// IsBotOwner returns whether or not a message sender was the owner of the bot.
func (d *Discord) IsBotOwner(message Message) bool {
	return message.UserID() == d.OwnerUserID
}

// IsPrivate returns whether or not a message was private.
func (d *Discord) IsPrivate(message Message) bool {
	c, err := d.Session.State.PrivateChannel(message.Channel())
	return err == nil && c.IsPrivate
}

// IsModerator returns whether or not the sender of a message is a moderator.
func (d *Discord) IsModerator(message Message) bool {
	c, err := d.Session.State.Channel(message.Channel())
	if err != nil {
		return false
	}
	g, err := d.Session.State.Guild(c.GuildID)
	if err != nil {
		return false
	}
	return g.OwnerID == message.UserID()
}

// ChannelCount returns the number of channels the bot is in.
func (d *Discord) ChannelCount() int {
	return len(d.Session.State.Guilds)
}

// SupportsMessageHistory returns if the service supports message history.
func (d *Discord) SupportsMessageHistory() bool {
	return true
}

// MessageHistory returns the message history for a channel.
func (d *Discord) MessageHistory(channel string) []Message {
	c, err := d.Session.State.Channel(channel)
	if err != nil {
		return nil
	}

	messages := make([]Message, len(c.Messages))
	for i := 0; i < len(c.Messages); i++ {
		messages[i] = &DiscordMessage{c.Messages[i], MessageTypeCreate}
	}

	return messages
}
