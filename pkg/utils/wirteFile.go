package utils

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

func WriteFile(filepath string, data interface{}) error {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 777)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer f.Close()

	switch data.(type) {
	case []byte:
		if _, err = f.Write(data.([]byte)); err != nil {
			logrus.Error(err)
			return err
		}
	default:
		jData, err := json.Marshal(data)
		if err != nil {
			logrus.Error(err)
			return err
		}
		if _, err = f.Write(jData); err != nil {
			logrus.Error(err)
			return err
		}
	}

	return nil
}
