package bucket

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStorage interface {
	UploadObject(
		ctx context.Context,
		fileKey string,
		file io.Reader,
		contentType string,
	) (string, error)
}

type objectStorageClient struct {
	Client *minio.Client
	name   string
}

func NewObjectStorageClient(
	endpoint, accessKeyID, secretAccessKey string,
	secure bool,
	name string,
) (*objectStorageClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, err
	}
	return &objectStorageClient{Client: client, name: name}, nil
}

func (o *objectStorageClient) GetOrCreateBucket(ctx context.Context) error {
	if o.name == "" {
		return nil
	}
	exists, err := o.Client.BucketExists(ctx, o.name)
	if err != nil {
		return err
	}
	if !exists {
		err = o.Client.MakeBucket(ctx, o.name, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *objectStorageClient) UploadObject(
	ctx context.Context,
	fileKey string,
	file io.Reader,
	contentType string,
) (string, error) {
	uploadinfo, err := o.Client.PutObject(
		ctx,
		o.name,
		fileKey,
		file,
		-1,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", err
	}
	return uploadinfo.Key, nil
}
