package uploader

import (
	"github.com/minio/minio-go/v6"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

type Uploader struct {
	client *minio.Client
	logger *zap.Logger
}

func NewUploaderOrDie(config Config, logger *zap.Logger) *Uploader {
	client, err := minio.New(config.Endpoint, config.AccessKey, config.SecretKey, config.SSL)
	if err != nil {
		panic(err)
	}
	return &Uploader{client: client, logger: logger}
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	SSL       bool
}

func (u *Uploader) Initialize(bucket string) error {
	return u.client.MakeBucket(bucket, "")
}

func (u *Uploader) Upload(bucket, keyPrefix, path string) error {
	filename := filepath.Base(path)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = u.client.PutObject(bucket, keyPrefix+"-"+filename, file, -1, minio.PutObjectOptions{})
	return err
}
