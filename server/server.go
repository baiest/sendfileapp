package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/baiest/sendfileapp/models"
	"github.com/baiest/sendfileapp/utils"
)

type Server struct {
	Host     string
	Port     int
	Channels map[models.ChannelId]*models.Channel
}

func NewServer(host string, port int) *Server {
	channels := make(map[models.ChannelId]*models.Channel)
	channels["1"] = models.NewChannel("1", make(models.ChannelStream))
	channels["2"] = models.NewChannel("2", make(models.ChannelStream))
	channels["3"] = models.NewChannel("3", make(models.ChannelStream))
	return &Server{
		Host:     host,
		Port:     port,
		Channels: channels,
	}
}

func (s *Server) Run() {
	log.Printf("Server is running in %s:%d", s.Host, s.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error: " + err.Error())
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 1024*10)
	_, err := conn.Read(buff)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error: ", err)
		}
	}

	req := utils.ToAction(buff)

	log.Println(fmt.Sprintf("{ Type: %s Channel: %s Filename: %s Data size: %d }", req.Type, req.ChannelId, req.FileName, len(req.Data)))
	s.reducer(req, conn)
}

func (s *Server) reducer(req *models.Action, conn net.Conn) {
	switch req.Type {
	case "received":
		s.clientReceive(conn, req)
	case "send":
		s.clientSend(conn, req)
	default:
		res := models.Action{
			Type: "log",
			Data: []byte("Acción no encontrada"),
		}
		resBuf := new(bytes.Buffer)
		gob.NewEncoder(resBuf).Encode(res)
		conn.Write(resBuf.Bytes())
	}
}

func (s *Server) clientReceive(conn net.Conn, req *models.Action) {
	res := models.Action{
		Type: "log",
		Data: nil,
	}
	resBuf := new(bytes.Buffer)
	channel, ok := s.Channels[models.ChannelId(req.ChannelId)]
	if !ok {
		res.Type = "close"
		res.Data = []byte(fmt.Sprintf("El canal '%s' no existe", req.ChannelId))
		gob.NewEncoder(resBuf).Encode(res)
		conn.Write(resBuf.Bytes())
		conn.Close()
		return
	}

	channel.AddClient(conn)
	res.Data = []byte(fmt.Sprintf("Agregado al canal %s", req.ChannelId))
	gob.NewEncoder(resBuf).Encode(res)
	conn.Write(resBuf.Bytes())

	for request := range channel.Stream {
		for _, client := range channel.Clients {
			res := models.Action{
				Type:      "file",
				FileName:  request.FileName,
				ChannelId: channel.Id,
				Data:      request.Data,
			}

			resBuf := new(bytes.Buffer)
			gob.NewEncoder(resBuf).Encode(res)
			_, err := client.Write(resBuf.Bytes())
			if err != nil {
				client.Close()
				channel.RemoveClient(client)
			}
		}
	}
}

func (s *Server) clientSend(conn net.Conn, req *models.Action) {
	channel := s.Channels[models.ChannelId(req.ChannelId)]
	channel.Stream <- req
}

func main() {
	server := NewServer("localhost", 3000)
	server.Run()
}
