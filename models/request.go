package models

type Request struct {
	Action    string
	ChannelId string
	Data      []byte
}
