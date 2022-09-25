package models

type Action struct {
	Type      string    `json:"type"`
	ChannelId ChannelId `json:"channel_id"`
	FileName  string    `json:"filename"`
	Data      []byte    `json:"data"`
}
