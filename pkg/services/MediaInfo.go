package services

import (
	"acourse-course-service/pkg/contracts"
	"io/ioutil"
	"log"
	"mime/multipart"
)

import "github.com/zhulik/go_mediainfo"

type MediaInfoService struct {
	mediainfo *mediainfo.MediaInfo
}

func (m MediaInfoService) GetVideoDuration(file *multipart.FileHeader) (string, error) {

	//Get raw file bytes
	openedFile, _ := file.Open()
	filebytes, err := ioutil.ReadAll(openedFile)
	if err != nil {
		return "", err
	}

	//Close file reading
	defer func(openedFile multipart.File) {
		err := openedFile.Close()
		if err != nil {
			log.Fatalf("Failed closing file, %v", err.Error())
		}
	}(openedFile)

	//Process Video
	err = m.mediainfo.OpenMemory(filebytes)
	if err != nil {
		return "", err
	}

	//panic("Duration >> " + m.mediainfo.Get("Duration"))
	return m.mediainfo.Get("Duration"), nil
}

func ConstructMediaInfoService(mediainfo *mediainfo.MediaInfo) contracts.MediaInfoService {
	return &MediaInfoService{mediainfo: mediainfo}
}
