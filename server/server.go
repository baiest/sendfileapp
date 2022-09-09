package server

import (
	"bytes"
	"encoding/gob"
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

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 1024)
	_, err := conn.Read(buff)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error: ", err)
		}
	}

	tmpBuff := bytes.NewBuffer(buff)
	req := new(models.Request)
	gob.NewDecoder(tmpBuff).Decode(req)
	s.reducer(req, conn)
}

func (s *Server) reducer(req *models.Request, conn net.Conn) {
	switch req.Action {
	case "received":
		s.clientReceive(conn, req)
	case "send":
		s.clientSend(conn, req)
	default:
		conn.Write([]byte("Acción no encontrada"))
	}
}

func (s *Server) clientReceive(conn net.Conn, req *models.Request) {
	channel, ok := s.Channels[models.ChannelId(req.ChannelId)]
	if !ok {
		conn.Write([]byte(fmt.Sprintf("El canal '%s' no existe", req.ChannelId)))
		conn.Close()
		return
	}
	channel.AddClient(conn)
	conn.Write([]byte("Agregado al channel:" + " " + string(req.ChannelId)))
	for data := range channel.Stream {
		for _, client := range channel.Clients {
			fmt.Println(client)
			client.Write(data)
		}
	}
}

func (s *Server) clientSend(conn net.Conn, req *models.Request) {
	channel := s.Channels[models.ChannelId(req.ChannelId)]
	channel.Stream <- req.Data
}
