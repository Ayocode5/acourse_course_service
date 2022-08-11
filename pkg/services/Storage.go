package services

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/response"
	"mime/multipart"
)

type StorageService struct {
	StorageRepository contracts.StorageRepository
}

func (s StorageService) UploadFile(file *multipart.FileHeader, prefix string) (response.S3Response, error) {
	res, err := s.StorageRepository.UploadFile(file, prefix)
	if err != nil {
		return response.S3Response{}, err
	}
	return res, nil
}

func (s StorageService) Delete(objectKey string) error {
	err := s.StorageRepository.DeleteObject(&objectKey)
	if err != nil {
		return err
	}
	return nil
}

func (s StorageService) UploadFiles(files []*multipart.FileHeader, prefix string) ([]response.S3Response, error) {

	res, err := s.StorageRepository.UploadFiles(files, prefix)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func ConstructStorageService(storageRepository *contracts.StorageRepository) contracts.StorageService {
	return &StorageService{StorageRepository: *storageRepository}
}
