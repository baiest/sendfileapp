package utils

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/baiest/sendfileapp/models"
)

func CreateFile(res *models.Action, wg *sync.WaitGroup, lock *sync.Mutex) {
	defer wg.Done()
	lock.Lock()
	defer lock.Unlock()
	err := os.MkdirAll(fmt.Sprintf("./channel-%s", res.ChannelId), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(fmt.Sprintf("./channel-%s/%s", res.ChannelId, res.FileName))
	if err != nil {
		log.Print(err)
	}
	_, err = file.Write(res.Data)
	if err != nil {
		log.Print(err)
	}
	file.Close()
	log.Printf("Archivo creado: '%s'", res.FileName)
}
