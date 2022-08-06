package contracts

import "mime/multipart"

type MediaInfoService interface {
	GetVideoDuration(file *multipart.FileHeader) (string, error)
}
