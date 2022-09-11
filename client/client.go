package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/baiest/sendfileapp/models"
	"github.com/baiest/sendfileapp/utils"
)

var channel = flag.String("c", "1", "channel")
var action = flag.String("a", "receive", "action")
var data = flag.String("d", "Hola", "data")
var filePath = flag.String("f", "", "file")

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

	err := os.MkdirAll(fmt.Sprintf("./channel-%s", req.ChannelId), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	conn, err := net.Dial("tcp", "localhost:3000")
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	if *action == "receive" {
		Receive(conn)
	} else {
		Send(conn)
		return
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
