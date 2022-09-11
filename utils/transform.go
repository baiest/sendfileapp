package utils

import (
	"bytes"
	"encoding/gob"

	"github.com/baiest/sendfileapp/models"
)

func ToAction(buff []byte) *models.Action {
	tmpBuff := bytes.NewBuffer(buff)
	action := new(models.Action)
	gob.NewDecoder(tmpBuff).Decode(action)
	return action
}
