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

		// inputMessage := bufio.NewScanner(conn)
		// for inputMessage.Scan() {
		s.readClient(conn, client)
		// 	fmt.Printf("%s: %s\n", client.Id, inputMessage.Text())
		// }
		// channel := s.Channels["1"]
		// client.Receive(channel)
		// fmt.Printf("%s:%d\n", client.Id, len(client.Channels))
	}
}

func (s *Server) readClient(conn net.Conn, client *models.Client) {
	buff := make([]byte, 1024)
	lenBuff, err := conn.Read(buff)
	tmpBuff := bytes.NewBuffer(buff[:lenBuff])
	req := new(models.Request)
	gob.NewDecoder(tmpBuff).Decode(req)

	s.reducer(req, client, conn)

	if err != nil {
		if err != io.EOF {
			fmt.Println("Error: ", err)
		}
	}
}

func (s *Server) reducer(req *models.Request, client *models.Client, conn net.Conn) {
	fmt.Println(req.Action)
	switch req.Action {
	case "received":
		client.Receive(s.Channels[models.ChannelId(req.Payload)])
		fmt.Println(client.Id, client.Channels)
		conn.Write([]byte("Agregado channel" + " " + string(req.Payload)))
	default:
		fmt.Println("Acción no encontrada")
	}
}
