package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	"github.com/baiest/sendfileapp/models"
	"github.com/baiest/sendfileapp/utils"
)

var channel = flag.String("channel", "1", "Channel's name to listen files")
var action = flag.String("action", "receive", "Action to send or receive files")
var filePath = flag.String("file", "", "Path's file to send")

func Send(conn net.Conn) {
	defer conn.Close()
	path, err := filepath.Abs(*filePath)
	if err != nil {
		log.Fatal(err)
	}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		conn.Close()
		log.Fatal(err)
		return
	}

	req := &models.Action{
		Type:      "send",
		ChannelId: models.ChannelId(*channel),
		FileName:  filepath.Base(path),
		Data:      file,
	}

	reqBuf := new(bytes.Buffer)
	gob.NewEncoder(reqBuf).Encode(req)
	conn.Write(reqBuf.Bytes())
}

func Receive(conn net.Conn) {

	req := &models.Action{
		Type:      "received",
		ChannelId: models.ChannelId(*channel),
	}

	bin_buf := new(bytes.Buffer)
	gob.NewEncoder(bin_buf).Encode(req)
	conn.Write(bin_buf.Bytes())
}

func main() {
	flag.Parse()

	conn, err := net.Dial("tcp", "localhost:3000")
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	switch *action {
	case "receive":
		Receive(conn)
	case "send":
		Send(conn)
	default:
		log.Fatal("Acción no encontrada, las acciones permitidas son: receive, send")
	}

	for {
		buff := make([]byte, 1024*20)
		_, err = conn.Read(buff)
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
		}

		res := utils.ToAction(buff)

		switch res.Type {
		case "file":
			log.Println("Reciviendo archivos...")
			utils.CreateFile(res)
			log.Printf("Archivo '%s' creado", res.FileName)
		case "log":
			log.Println(string(res.Data))
		case "close":
			log.Println(string(res.Data))
			conn.Close()
			return
		}
	}
}
