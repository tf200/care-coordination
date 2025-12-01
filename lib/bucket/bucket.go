package bucket

import "github.com/minio/minio-go/v7"

type ObjectStorage interface {
}

type objectStorageClient struct {
	Client *minio.Client
}

func NewObjectStorageClient(endpoint, accessKeyID, secretAccessKey string, secure bool) (*objectStorageClient, error) {
	client, err := minio.New(endpoint, &minio.Options{})
}