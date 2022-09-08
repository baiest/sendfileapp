package models

type ChannelId string
type ChannelStream chan<- interface{}

type Channel struct {
	Id     ChannelId
	Stream ChannelStream
}

func NewChannel(id ChannelId, channel ChannelStream) *Channel {
	return &Channel{
		Id:     id,
		Stream: channel,
	}
}
