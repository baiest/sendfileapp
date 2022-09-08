package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/baiest/sendfileapp/server/models"
)

type Server struct {
	Host     string
	Port     int
	Channels map[models.ChannelId]models.Channel
	Clients  []models.Client
}

func NewServer(host string, port int) *Server {
	channels := make(map[models.ChannelId]models.Channel)
	channels["1"] = *models.NewChannel("1", make(models.ChannelStream))
	channels["2"] = *models.NewChannel("2", make(models.ChannelStream))
	channels["3"] = *models.NewChannel("3", make(models.ChannelStream))
	return &Server{
		Host:     host,
		Port:     port,
		Channels: channels,
		Clients:  make([]models.Client, 0),
	}
}

func (s *Server) Run() {
	log.Printf("Server is running in %s:%d", s.Host, s.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		log.Fatal(err)
	}

	log.Println(s.Channels)
	for {
		conn, err := listener.Accept()
		defer conn.Close()
		if err != nil {
			log.Println("Error: " + err.Error())
			continue
		}
		client := models.NewClient(conn.RemoteAddr().String())
		s.readClient(conn, client)
	}
}

func (s *Server) readClient(conn net.Conn, client *models.Client) {
	buff := make([]byte, 1024)
	_, err := conn.Read(buff)
	fmt.Println("Cliente conectado")
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error: ", err)
		}
	}

	tmpBuff := bytes.NewBuffer(buff)
	req := new(models.Request)
	gob.NewDecoder(tmpBuff).Decode(req)

	s.reducer(req, client, conn)
}

func (s *Server) reducer(req *models.Request, client *models.Client, conn net.Conn) {
	fmt.Println(req.Action)
	switch req.Action {
	case "received":
		go func() {
			channel := s.Channels[models.ChannelId(req.ChannelId)]
			client.Receive(channel)
			fmt.Println(client.Id, client.Channels, channel.Id)
			conn.Write([]byte("Agregado channel" + " " + string(req.ChannelId) + "\n"))
			for {
				for data := range channel.Stream {
					conn.Write(data)
					conn.Write([]byte("\n"))
				}
			}
		}()
	case "send":
		go func() {
			channel := s.Channels[models.ChannelId(req.ChannelId)]
			channel.Stream <- req.Data
		}()
	default:
		fmt.Println("Acción no encontrada")
	}
}
