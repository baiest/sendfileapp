package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/baiest/sendfileapp/models"
)

func CreateFile(res *models.Action) {
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
}
