package uploader

import (
	"go.uber.org/zap"
	"io/ioutil"
)

func Upload(logger *zap.Logger, path string) error {
	// TODO prints out contents for now
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	logger.Info(string(file))
	return nil
}
