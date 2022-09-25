package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"sync"

	"github.com/baiest/sendfileapp/models"
	"github.com/baiest/sendfileapp/utils"
)

/*
	Flags to command line
*/
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
	data, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(len(data))
	conn.Write(data)
}

func Receive(conn net.Conn) {

	req := &models.Action{
		Type:      "received",
		ChannelId: models.ChannelId(*channel),
	}
	data, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}
	conn.Write(data)
}

func readData(buff []byte, wg *sync.WaitGroup, lock *sync.Mutex) {
	var res models.Action
	err := json.Unmarshal(bytes.Trim(buff, "\x00"), &res)
	if err != nil {
		log.Println("Reciviendo respuesta:", err)
		return
	}

	//Manage type of data
	switch res.Type {
	case "file":
		log.Println("Reciviendo archivos...")
		wg.Add(1)
		go utils.CreateFile(&res, wg, lock)
	case "log":
		log.Println(string(res.Data))
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

	//Manage flag to action
	switch *action {
	case "receive":
		Receive(conn)
	case "send":
		Send(conn)
		return
	default:
		log.Fatal("Acción no encontrada, las acciones permitidas son: receive, send")
	}

	wg := &sync.WaitGroup{}
	lock := &sync.Mutex{}

	const BUFFER_SIZE = 256
	buff := make([]byte, BUFFER_SIZE)
	var totalBuff []byte

	//Listen response from server
	for {
		n, err := conn.Read(buff)
		totalBuff = append(totalBuff, buff[:n]...)
		if err != nil {
			if err == io.EOF {
				log.Println("Terminado de leer")
			} else {
				log.Fatal(err)
			}
		}
		//Full buffer with data
		if n < BUFFER_SIZE {
			readData(totalBuff, wg, lock)
			buff = make([]byte, BUFFER_SIZE)
			totalBuff = []byte{}

			wg.Wait()
		}
	}
}
