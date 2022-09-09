package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"io"
	"log"
	"net"

	"github.com/baiest/sendfileapp/models"
)

var channel = flag.String("c", "1", "channel")
var action = flag.String("a", "receive", "action")
var data = flag.String("d", "Hola", "data")

func Connect() {
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
	done := make(chan struct{})
	go func() {
		io.Copy(log.Writer(), conn)
		done <- struct{}{}
	}()
	<-done
}

func Send(conn net.Conn) {
	defer conn.Close()

	req := &models.Request{
		Action:    "send",
		ChannelId: *channel,
		Data:      []byte(*data),
	}

	reqBuf := new(bytes.Buffer)
	gob.NewEncoder(reqBuf).Encode(req)
	conn.Write(reqBuf.Bytes())
}

func Receive(conn net.Conn) {
	req := &models.Request{
		Action:    "received",
		ChannelId: *channel,
	}
	bin_buf := new(bytes.Buffer)
	gob.NewEncoder(bin_buf).Encode(req)
	conn.Write(bin_buf.Bytes())
}

func main() {
	flag.Parse()
	Connect()
}
