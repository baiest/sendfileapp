package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/baiest/sendfileapp/models"
)

func CreateFile(res *models.Action) {
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
