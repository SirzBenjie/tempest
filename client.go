package ashara

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"
)

// Client is the core Ashara entrypoint
type Client struct {
	ApplicationID   Snowflake
	PublicKey       ed25519.PublicKey
	Rest            RestHandler
	CommandRegistry SlashCommandRegistry

	jsonBufferPool *sync.Pool
}

type ClientOptions struct {
	Token          string
	PublicKey      string
	JSONBufferSize uint
}

func NewClient(opt ClientOptions) Client {
	discordPublicKey, err := hex.DecodeString(opt.PublicKey)
	if err != nil {
		panic("failed to decode discord's public key (check if it's correct key): " + err.Error())
	}

	botUserID, err := extractUserIDFromToken(opt.Token)
	if err != nil {
		panic("failed to extract bot user ID from bot token: " + err.Error())
	}

	var poolSize uint = 4096
	if opt.JSONBufferSize > poolSize {
		poolSize = opt.JSONBufferSize
	}

	return Client{
		ApplicationID:   botUserID,
		PublicKey:       discordPublicKey,
		Rest:            NewBaseRestHandler(opt.Token),
		CommandRegistry: NewBaseSlashCommandRegistry(botUserID),
		jsonBufferPool: &sync.Pool{
			New: func() any {
				buf := make([]byte, poolSize) // start with a decent buffer
				return &buf
			},
		},
	}
}

// Pings Discord API and returns time it took to get response.
func (client *Client) Ping() time.Duration {
	start := time.Now()
	client.Rest.Request(http.MethodGet, "/gateway", nil)
	return time.Since(start)
}

func (client *Client) SendMessage(channelID Snowflake, message Message, files []*os.File) (Message, error) {
	raw, err := client.Rest.RequestWithFiles(http.MethodPost, "/channels/"+channelID.String()+"/messages", message, files)
	if err != nil {
		return Message{}, err
	}

	res := Message{}
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return Message{}, errors.New("failed to parse received data from discord")
	}

	return res, nil
}

func (client *Client) SendLinearMessage(channelID Snowflake, content string) (Message, error) {
	return client.SendMessage(channelID, Message{Content: content}, nil)
}

// Creates (or fetches if already exists) user's private text channel (DM) and tries to send message into it.
// Warning! Discord's user channels endpoint has huge rate limits so please reuse Message#ChannelID whenever possible.
func (client *Client) SendPrivateMessage(userID Snowflake, content Message, files []*os.File) (Message, error) {
	res := make(map[string]interface{}, 0)
	res["recipient_id"] = userID

	raw, err := client.Rest.Request(http.MethodPost, "/users/@me/channels", res)
	if err != nil {
		return Message{}, err
	}

	err = json.Unmarshal(raw, &res)
	if err != nil {
		return Message{}, errors.New("failed to parse received data from discord")
	}

	channelID, err := StringToSnowflake(res["id"].(string))
	if err != nil {
		return Message{}, err
	}

	msg, err := client.SendMessage(channelID, content, files)
	msg.ChannelID = channelID // Just in case.

	return msg, err
}

func (client *Client) EditMessage(channelID Snowflake, messageID Snowflake, content Message) error {
	_, err := client.Rest.Request(http.MethodPatch, "/channels/"+channelID.String()+"/messages/"+messageID.String(), content)
	return err
}

func (client *Client) DeleteMessage(channelID Snowflake, messageID Snowflake) error {
	_, err := client.Rest.Request(http.MethodDelete, "/channels/"+channelID.String()+"/messages/"+messageID.String(), nil)
	return err
}

func (client *Client) CrosspostMessage(channelID Snowflake, messageID Snowflake) error {
	_, err := client.Rest.Request(http.MethodPost, "/channels/"+channelID.String()+"/messages/"+messageID.String()+"/crosspost", nil)
	return err
}

func (client *Client) FetchUser(id Snowflake) (User, error) {
	raw, err := client.Rest.Request(http.MethodGet, "/users/"+id.String(), nil)
	if err != nil {
		return User{}, err
	}

	res := User{}
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return User{}, errors.New("failed to parse received data from discord")
	}

	return res, nil
}

func (client *Client) FetchMember(guildID Snowflake, memberID Snowflake) (Member, error) {
	raw, err := client.Rest.Request(http.MethodGet, "/guilds/"+guildID.String()+"/members/"+memberID.String(), nil)
	if err != nil {
		return Member{}, err
	}

	res := Member{}
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return Member{}, errors.New("failed to parse received data from discord")
	}

	return res, nil
}
