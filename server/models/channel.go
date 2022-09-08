package models

type ChannelId string
type ChannelStream chan []byte
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
