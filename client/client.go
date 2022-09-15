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
	"path/filepath"
	"sync"

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
	_, err = conn.Write(reqBuf.Bytes())
	if err != nil {
		log.Println("Error encoding", err)
	}
	fmt.Println(len(reqBuf.Bytes()))
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
		return
	default:
		log.Fatal("Acción no encontrada, las acciones permitidas son: receive, send")
	}

	wg := &sync.WaitGroup{}
	lock := &sync.Mutex{}

	const BUFFER_SIZE = 256
	buff := make([]byte, BUFFER_SIZE)
	var totalBuff []byte
	for {
		n, err := conn.Read(buff)
		totalBuff = append(totalBuff, buff...)
		if err != nil {
			if err == io.EOF {
				log.Println("Terminado de leer")
				buff = make([]byte, 1024)
				totalBuff = []byte{}
			} else {
				log.Fatal(err)
			}
		}
		if n < BUFFER_SIZE {
			res := utils.ToAction(totalBuff)
			switch res.Type {
			case "file":
				log.Println("Reciviendo archivos...")
				wg.Add(1)
				go utils.CreateFile(res, wg, lock)
			case "log":
				log.Println(string(res.Data))
			}
			totalBuff = []byte{}

			wg.Wait()
		}
	}
}

// for {
// 	for {
// 		_, err := conn.Read(buff)
// 		totalBuff = append(totalBuff, buff...)
// 		if err != nil {
// 			if err == io.EOF {
// 				log.Println("Terminado de leer")
// 				res := utils.ToAction(totalBuff)

// 				wg := &sync.WaitGroup{}
// 				lock := &sync.Mutex{}

// 				switch res.Type {
// 				case "file":
// 					log.Println("Reciviendo archivos...")
// 					wg.Add(1)
// 					go utils.CreateFile(res, wg, lock)
// 					log.Printf("Archivo '%s' creado", res.FileName)
// 				case "log":
// 					log.Println(string(res.Data))
// 				}
// 				wg.Wait()
// 				break
// 			}
// 			log.Fatal(err)
// 		}
// 	}
// 	log.Println("Leyendo...", len(totalBuff))
// }

// res := utils.ToAction(buff)

// wg := &sync.WaitGroup{}
// lock := &sync.Mutex{}

// switch res.Type {
// case "file":
// 	log.Println("Reciviendo archivos...")
// 	wg.Add(1)
// 	go utils.CreateFile(res, wg, lock)
// 	log.Printf("Archivo '%s' creado", res.FileName)
// case "log":
// 	log.Println(string(res.Data))
// 	// case "close":
// 	// 	log.Println(string(res.Data))
// 	// 	conn.Close()
// 	// 	return
// }
// wg.Wait()
