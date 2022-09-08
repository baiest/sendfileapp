package models

type Client struct {
	Id       string
	Channels map[ChannelId]Channel
}

func NewClient(id string) *Client {
	return &Client{
		Id:       id,
		Channels: make(map[ChannelId]Channel),
	}
}

func (c *Client) Receive(channel Channel) {
	c.Channels[channel.Id] = channel
}
