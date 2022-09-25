package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/baiest/sendfileapp/models"
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

func (s *Server) readData(buff []byte, conn net.Conn) {
	var req models.Action
	totalBuff := bytes.Trim(buff, "\x00")
	err := json.Unmarshal(totalBuff, &req)
	if err != nil {
		log.Println("Error", err)
	}
	log.Println(fmt.Sprintf("{ Type: %s Channel: %s Filename: %s Data size: %d }", req.Type, req.ChannelId, req.FileName, len(req.Data)))
	go s.reducer(&req, conn)
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	const BUFFER_SIZE = 256
	buff := make([]byte, BUFFER_SIZE)
	var totalBuff []byte

	//Listen requests from clients
	for {
		n, err := conn.Read(buff)
		totalBuff = append(totalBuff, buff[:n]...)
		if err != nil {
			if err == io.EOF {
				log.Println("finish.")
			} else {
				log.Println(err)
			}
			for _, channel := range s.Channels {
				channel.RemoveClient(conn)
			}
			break
		}
		// Full buffer with data
		if n < BUFFER_SIZE {
			s.readData(totalBuff, conn)
		}
	}
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
		data, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
		}
		conn.Write(data)
	}
}

func (s *Server) clientReceive(conn net.Conn, req *models.Action) {
	res := models.Action{
		Type: "log",
	}
	channel, ok := s.Channels[models.ChannelId(req.ChannelId)]
	if !ok {
		res.Type = "close"
		res.Data = []byte(fmt.Sprintf("El canal '%s' no existe", req.ChannelId))
		data, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
		}
		conn.Write(data)
		conn.Close()
		return
	}

	channel.AddClient(conn)
	res.Data = []byte(fmt.Sprintf("Agregado al canal %s", req.ChannelId))
	data, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
	}
	conn.Write(data)

	for request := range channel.Stream {
		res := models.Action{
			Type:      "file",
			FileName:  request.FileName,
			ChannelId: channel.Id,
			Data:      request.Data,
		}
		data, err := json.Marshal(res)
		for _, client := range channel.Clients {
			if err != nil {
				log.Println(err)
				client.Close()
				channel.RemoveClient(client)
			}
			client.Write(data)
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
