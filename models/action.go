package models

type Action struct {
	Type      string
	ChannelId ChannelId
	FileName  string
	Data      []byte
}
