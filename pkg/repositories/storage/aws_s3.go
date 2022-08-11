package repositories

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/response"
	"bytes"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/exp/slices"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
	"time"
)

type S3BucketService struct {
	maxPartSize     int64
	maxRetries      int
	accessKeyID     string
	secretAccessKey string
	bucketRegion    string
	bucketName      string
	allowedMimeType []string
	session         *session.Session
	prefix          *string
}

func ConstructS3Repository(accessKeyId string, secretAccessKey string, bucketName string, bucketRegion string, allowedMimeType []string, maxPartSize int64, maxRetries int) contracts.StorageRepository {
	return &S3BucketService{
		maxPartSize:     maxPartSize,
		maxRetries:      maxRetries,
		accessKeyID:     accessKeyId,
		secretAccessKey: secretAccessKey,
		bucketRegion:    bucketRegion,
		bucketName:      bucketName,
		allowedMimeType: allowedMimeType,
	}
}

func (s S3BucketService) GetClient() (*s3.S3, error) {

	var client *s3.S3

	credential := credentials.NewStaticCredentials(s.accessKeyID, s.secretAccessKey, "")
	_, err := credential.Get()
	if err != nil {
		return nil, err
	}

	cfg := aws.NewConfig().WithRegion(s.bucketRegion).WithCredentials(credential)
	newSession, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	if client == nil {
		return s3.New(newSession, cfg), nil
	}

	return client, nil

}

//Read file bytes From multipart request
func (s S3BucketService) readFileBytes(file *multipart.FileHeader) ([]byte, error) {
	//Get raw file bytes
	openedFile, _ := file.Open()
	filebytes, err := ioutil.ReadAll(openedFile)

	if err != nil {
		return nil, err
	}

	//Close file reading
	defer func(openedFile multipart.File) {
		err := openedFile.Close()
		if err != nil {
			log.Fatalf("Failed closing file, %v", err.Error())
		}
	}(openedFile)

	return filebytes, nil
}

//Files must be video type
func (s S3BucketService) validateFilesType(files []*multipart.FileHeader) (bool, error) {

	for _, file := range files {
		size := file.Size
		contentType := file.Header.Get("Content-Type")

		if size > s.maxPartSize {
			return false, errors.New("file too large")
		}

		if !slices.Contains(s.allowedMimeType, contentType) {
			return false, errors.New(fmt.Sprintf("File type %v is not supported", file.Filename))
		}
	}

	return true, nil
}

func (s S3BucketService) validateFileType(file *multipart.FileHeader) (bool, error) {

	size := file.Size
	contentType := file.Header.Get("Content-Type")

	if size > s.maxPartSize {
		return false, errors.New("file too large")
	}

	if !slices.Contains(s.allowedMimeType, contentType) {
		return false, errors.New(fmt.Sprintf("File type %v is not supported", file.Filename))
	}

	return true, nil
}

//Construct AWS CompleteMultipartUpload Object
func (s S3BucketService) completeMultipartUpload(S3 *s3.S3, resp *s3.CreateMultipartUploadOutput, completedParts []*s3.CompletedPart) (*s3.CompleteMultipartUploadOutput, error) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}

	return S3.CompleteMultipartUpload(completeInput)
}

//Construct AWS CompletedPart
func (s S3BucketService) uploadPart(S3 *s3.S3, resp *s3.CreateMultipartUploadOutput, filebytes []byte, partNumber int) (*s3.CompletedPart, error) {

	try := 1

	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(filebytes),
		Bucket:        resp.Bucket,
		Key:           resp.Key,
		PartNumber:    aws.Int64(int64(partNumber)),
		UploadId:      resp.UploadId,
		ContentLength: aws.Int64(int64(len(filebytes))),
	}

	for try <= s.maxRetries {
		uploadPartOutput, err := S3.UploadPart(partInput)
		if err != nil {
			if try == s.maxRetries {
				if aerr, ok := err.(awserr.Error); ok {
					return nil, aerr
				}
				return nil, err
			}
			log.Printf("Retrying to upload part #%v\n", partNumber)
			try++
		} else {
			//fmt.Printf("UploadFiles part #%v\n", partNumber)
			completed := &s3.CompletedPart{
				ETag:       uploadPartOutput.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}

			return completed, nil
		}
	}

	return nil, nil
}

//Abort multipart If one of the upload part process failed
func (s S3BucketService) abortMultiPartUpload(S3 *s3.S3, resp *s3.CreateMultipartUploadOutput) error {
	log.Println("Aborting multipart upload for UploadId#" + *resp.UploadId)
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
	}
	_, err := S3.AbortMultipartUpload(abortInput)
	return err
}

func (s S3BucketService) UploadFile(file *multipart.FileHeader, prefix string) (response.S3Response, error) {
	//1. Validate Inputs
	//valid, err := s.validateFileType(file)
	//if !valid {
	//	return response.S3Response{}, err
	//}

	//2. Open AWS S3 Session
	s3Client, _ := s.GetClient()

	//3.
	now := time.Now()
	nowRFC3339 := now.Format(time.RFC3339)

	//4. Loop over form input and process each part
	var wg sync.WaitGroup
	finalResultChannel := make(chan response.S3Response, 10)
	pathNumber := 0

	wg.Add(1)

	//5. Get filePart bytes
	go func(wg *sync.WaitGroup, filePart *multipart.FileHeader, pathNumber int, prefix string) {

		defer wg.Done()

		result := response.S3Response{}

		filebytes, err := s.readFileBytes(filePart)
		if err != nil {
			log.Println(err.Error())
			result.Success = false
			result.Filename = filePart.Filename
			result.Message = err.Error()
			result.Order = pathNumber
			result.Key = ""
			finalResultChannel <- result
			return
		}

		//6. Filename format
		path := prefix + nowRFC3339 + "-" + filePart.Filename

		//7. Get filePart type /mime type
		fileType := http.DetectContentType(filebytes)

		//8. Prepare s3 multipart upload
		input := &s3.CreateMultipartUploadInput{
			Bucket:      aws.String(s.bucketName),
			Key:         aws.String(path),
			ContentType: aws.String(fileType),
		}

		//9. Create s3 multipart upload
		createdMultipartOutput, err := s3Client.CreateMultipartUpload(input)
		if err != nil {
			log.Println(err.Error())
			result.Success = false
			result.Filename = filePart.Filename
			result.Order = pathNumber
			result.Message = err.Error()
			result.Key = ""
			finalResultChannel <- result
			return
		}

		log.Println("Created multipart upload request")

		//10. UploadFiles Multipart
		var current, partLength int64
		var remaining = filePart.Size
		var completedParts []*s3.CompletedPart

		partNumber := 1
		for current = 0; remaining != 0; current += partLength {
			if remaining < s.maxPartSize {
				partLength = remaining
			} else {
				partLength = s.maxPartSize
			}

			//UploadFiles Binaries File
			completedPart, err := s.uploadPart(s3Client, createdMultipartOutput, filebytes[current:current+partLength], partNumber)

			if err != nil {
				log.Println(err.Error())

				err := s.abortMultiPartUpload(s3Client, createdMultipartOutput)
				if err != nil {
					log.Println(err.Error())
					result.Success = false
					result.Filename = filePart.Filename
					result.Order = pathNumber
					result.Message = err.Error()
					result.Key = ""
					finalResultChannel <- result

					return
				}

				result.Success = false
				result.Filename = filePart.Filename
				result.Order = pathNumber
				result.Message = err.Error()
				result.Key = ""
				finalResultChannel <- result

				return
			}

			remaining -= partLength
			partNumber++
			completedParts = append(completedParts, completedPart)

		}

		log.Printf("Uploading file: %s\n", filePart.Filename)

		completeResponse, err := s.completeMultipartUpload(s3Client, createdMultipartOutput, completedParts)

		if err != nil {
			log.Println(err.Error())
			result.Success = false
			result.Filename = filePart.Filename
			result.Order = pathNumber
			result.Message = err.Error()
			result.Key = ""
			finalResultChannel <- result
			return
		}

		log.Printf("File successfully uploaded : %s\n", filePart.Filename)

		//Array of uploaded part's url location
		result.Success = true
		result.Filename = filePart.Filename
		result.Filepath = *completeResponse.Location
		result.Order = pathNumber
		result.Key = *completeResponse.Key
		result.Message = fmt.Sprintf("File %v successfully uploaded", filePart.Filename)
		finalResultChannel <- result

	}(&wg, file, pathNumber, prefix)

	pathNumber++

	wg.Wait()

	//Close Channel
	close(finalResultChannel)

	//Loop Over Channel

	return <-finalResultChannel, nil
}

func (s S3BucketService) UploadFiles(files []*multipart.FileHeader, prefix string) ([]response.S3Response, error) {

	//1. Validate Inputs
	//valid, err := s.validateFilesType(files)
	//if !valid {
	//	return nil, err
	//}

	//2. Open AWS S3 Session
	s3Client, _ := s.GetClient()

	//3.
	now := time.Now()
	nowRFC3339 := now.Format(time.RFC3339)

	//4. Loop over form input and process each part
	var wg sync.WaitGroup
	finalResultChannel := make(chan response.S3Response, 10)
	pathNumber := 0

	for _, filePart := range files {

		wg.Add(1)

		//5. Get filePart bytes
		go func(wg *sync.WaitGroup, filePart *multipart.FileHeader, pathNumber int, prefix string) {

			defer wg.Done()

			result := response.S3Response{}

			filebytes, err := s.readFileBytes(filePart)
			if err != nil {
				log.Println(err.Error())
				result.Success = false
				result.Filename = filePart.Filename
				result.Message = err.Error()
				result.Order = pathNumber
				result.Key = ""
				finalResultChannel <- result
				return
			}

			//6. Filename format
			path := prefix + nowRFC3339 + "-" + filePart.Filename

			//7. Get filePart type /mime type
			fileType := http.DetectContentType(filebytes)

			//8. Prepare s3 multipart upload
			input := &s3.CreateMultipartUploadInput{
				Bucket:      aws.String(s.bucketName),
				Key:         aws.String(path),
				ContentType: aws.String(fileType),
			}

			//9. Create s3 multipart upload
			createdMultipartOutput, err := s3Client.CreateMultipartUpload(input)
			if err != nil {
				log.Println(err.Error())
				result.Success = false
				result.Filename = filePart.Filename
				result.Order = pathNumber
				result.Message = err.Error()
				result.Key = ""
				finalResultChannel <- result
				return
			}

			log.Println("Created multipart upload request")

			//10. UploadFiles Multipart
			var current, partLength int64
			var remaining = filePart.Size
			var completedParts []*s3.CompletedPart

			partNumber := 1
			for current = 0; remaining != 0; current += partLength {

				if remaining < s.maxPartSize {
					partLength = remaining
				} else {
					partLength = s.maxPartSize
				}

				//UploadFiles Binaries File
				completedPart, err := s.uploadPart(s3Client, createdMultipartOutput, filebytes[current:current+partLength], partNumber)

				if err != nil {
					log.Println(err.Error())

					err := s.abortMultiPartUpload(s3Client, createdMultipartOutput)
					if err != nil {
						log.Println(err.Error())
						result.Success = false
						result.Filename = filePart.Filename
						result.Order = pathNumber
						result.Message = err.Error()
						result.Key = ""
						finalResultChannel <- result

						return
					}

					result.Success = false
					result.Filename = filePart.Filename
					result.Order = pathNumber
					result.Message = err.Error()
					result.Key = ""
					finalResultChannel <- result

					return
				}

				remaining -= partLength
				partNumber++
				completedParts = append(completedParts, completedPart)

			}

			log.Printf("Uploading file: %s\n", filePart.Filename)

			completeResponse, err := s.completeMultipartUpload(s3Client, createdMultipartOutput, completedParts)

			if err != nil {
				log.Println(err.Error())
				result.Success = false
				result.Filename = filePart.Filename
				result.Order = pathNumber
				result.Message = err.Error()
				result.Key = ""
				finalResultChannel <- result
				return
			}

			log.Printf("File successfully uploaded : %s\n", filePart.Filename)

			//Array of uploaded part's url location
			result.Success = true
			result.Filename = filePart.Filename
			result.Filepath = *completeResponse.Location
			result.Order = pathNumber
			result.Key = *completeResponse.Key
			result.Message = fmt.Sprintf("File %v successfully uploaded", filePart.Filename)
			finalResultChannel <- result

		}(&wg, filePart, pathNumber, prefix)

		pathNumber++
	}

	wg.Wait()

	//Close Channel
	close(finalResultChannel)

	//Loop Over Channel
	var finalResult []response.S3Response

	for res := range finalResultChannel {
		finalResult = append(finalResult, res)
	}

	return finalResult, nil

}

func (s S3BucketService) DeleteObject(objectKey *string) error {

	client, err := s.GetClient()
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    objectKey,
	})
	if err != nil {
		return err
	}

	err = client.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    objectKey,
	})
	if err != nil {
		return err
	}

	return nil
}
