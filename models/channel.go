package models

import (
	"fmt"
	"net"
)

type ChannelId string
type ChannelStream chan []byte
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
	fmt.Println(c.Clients)
}
