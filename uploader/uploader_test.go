package uploader

import (
	"errors"
	"github.com/minio/minio-go/v6"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"jobber/test/check"
	"path/filepath"
	"testing"
)

func TestUploader_Initialize(t *testing.T) {
	fakeClient := &stubbedClient{}
	uploader := Uploader{
		client: fakeClient,
		logger: zap.L(),
	}

	err := uploader.Initialize("bucketName")
	check.Ok(t, err)
	check.Equals(t, "bucketName", fakeClient.bucketName)
}

func TestUploader_InitializeReturnsError(t *testing.T) {
	fakeClient := &stubbedClient{err: errors.New("failed")}
	uploader := Uploader{
		client: fakeClient,
		logger: zap.L(),
	}

	err := uploader.Initialize("bucketName")
	check.Assert(t, err != nil, "expected error not to be nil")
	check.Equals(t, "", fakeClient.bucketName)
}

func TestUploader_Upload(t *testing.T) {
	file, _ := ioutil.TempFile("", "example.*.txt")
	_, _ = file.WriteString("hello")
	filename := filepath.Base(file.Name())
	defer file.Close()
	fakeClient := &stubbedClient{}
	uploader := Uploader{
		client: fakeClient,
		logger: zap.L(),
	}

	err := uploader.Upload("bucketName", "prefix", file.Name())

	check.Ok(t, err)
	check.Equals(t, "bucketName", fakeClient.bucketName)
	check.Equals(t, "prefix-"+filename, fakeClient.objectName)

	check.Equals(t, "hello", string(fakeClient.read))
}

func TestUploader_Fails(t *testing.T) {
	file, _ := ioutil.TempFile("", "example.*.txt")
	_, _ = file.WriteString("hello")
	defer file.Close()
	fakeClient := &stubbedClient{err: errors.New("failed")}
	uploader := Uploader{
		client: fakeClient,
		logger: zap.L(),
	}

	err := uploader.Upload("bucketName", "prefix", file.Name())

	check.Assert(t, err != nil, "expected error not to be nil")
	check.Equals(t, "", fakeClient.bucketName)
	check.Equals(t, "", fakeClient.objectName)
	check.Equals(t, "", string(fakeClient.read))
}

type stubbedClient struct {
	err        error
	bucketName string
	location   string
	objectName string
	read       []byte
}

func (f *stubbedClient) PutObject(bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	f.bucketName = bucketName
	f.objectName = objectName
	f.read, _ = ioutil.ReadAll(reader)
	return 0, nil
}

func (f *stubbedClient) MakeBucket(bucketName string, location string) error {
	if f.err != nil {
		return f.err
	}
	f.bucketName = bucketName
	f.location = location
	return nil
}
