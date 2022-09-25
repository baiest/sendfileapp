package models

import (
	"net"
)

type ChannelId string
type ChannelStream chan *Action
type Channel struct {
	Id      ChannelId
	Stream  ChannelStream
	Clients []net.Conn
}

func NewChannel(id ChannelId, channel ChannelStream) *Channel {
	return &Channel{
		Id:      id,
		Stream:  channel,
		Clients: make([]net.Conn, 0),
	}
}

func (c *Channel) AddClient(client net.Conn) {
	c.Clients = append(c.Clients, client)
}

func (c *Channel) RemoveClient(client net.Conn) {
	for index, current := range c.Clients {
		if current == client {
			c.Clients = append(c.Clients[:index], c.Clients[index+1:]...)
			break
		}
	}
}
