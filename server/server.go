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
	const BUFFER_SIZE = 256
	buff := make([]byte, BUFFER_SIZE)
	var totalBuff []byte
	for {
		n, err := conn.Read(buff)
		totalBuff = append(totalBuff, buff...)
		if err != nil {
			if err == io.EOF {
				log.Println("finish.")
				break
			}
			fmt.Println("Error: ", err)
			for _, channel := range s.Channels {
				channel.RemoveClient(conn)
				log.Println(channel.Clients)
			}
			break
		}
		if n < BUFFER_SIZE {
			req := utils.ToAction(totalBuff)
			log.Println(fmt.Sprintf("{ Type: %s Channel: %s Filename: %s Data size: %d }", req.Type, req.ChannelId, req.FileName, len(req.Data)))
			go s.reducer(req, conn)
			totalBuff = []byte{}
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
	log.Println("Terminado de subir")
	channel := s.Channels[models.ChannelId(req.ChannelId)]
	channel.Stream <- req
}

func main() {
	server := NewServer("localhost", 3000)
	server.Run()
}
