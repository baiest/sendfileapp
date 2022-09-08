package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"io"
	"log"
	"net"
	"os"

	"github.com/baiest/sendfileapp/client/models"
)

var channel = flag.String("c", "1", "channel")
var action = flag.String("a", "receive", "action")

func Connect() {
	conn, err := net.Dial("tcp", "localhost:3000")
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	req := &models.Request{
		Action:  "received",
		Payload: []byte(*channel),
	}
	bin_buf := new(bytes.Buffer)
	gob.NewEncoder(bin_buf).Encode(req)
	conn.Write(bin_buf.Bytes())

	done := make(chan struct{})
	go func() {
		io.Copy(os.Stdout, conn)
		done <- struct{}{}
	}()
	CopyContent(conn, os.Stdin)
	conn.Close()
	<-done
}

func CopyContent(dst io.Writer, src io.Reader) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Fatal(err)
	}
}

func Send() {
	panic("Hola Implementa send")
}

func main() {
	flag.Parse()
	if true {
		Connect()
	} else {
		Send()
	}
}
