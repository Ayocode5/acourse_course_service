package contracts

import (
	"acourse-course-service/pkg/http/response"
	"github.com/aws/aws-sdk-go/service/s3"
	"mime/multipart"
)

type StorageRepository interface {
	UploadFiles(files []*multipart.FileHeader, prefix string) ([]response.S3Response, error)
	UploadFile(file *multipart.FileHeader, prefix string) (response.S3Response, error)
	DeleteObject(objectKey *string) error
	GetClient() (*s3.S3, error)
}

type StorageService interface {
	UploadFiles(files []*multipart.FileHeader, prefix string) ([]response.S3Response, error)
	UploadFile(file *multipart.FileHeader, prefix string) (response.S3Response, error)
	Delete(objectKey string) error
}
