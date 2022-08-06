package services

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/requests"
	"acourse-course-service/pkg/http/response"
	"acourse-course-service/pkg/models"
	"errors"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

type CourseService struct {
	DBRepository     contracts.CourseDatabaseRepository
	StorageService   contracts.StorageService
	MediaInfoService contracts.MediaInfoService
}

func ConstructCourseService(
	dbRepository *contracts.CourseDatabaseRepository,
	storageService *contracts.StorageService,
	mediaInfoService *contracts.MediaInfoService) contracts.CourseService {

	return &CourseService{
		DBRepository:     *dbRepository,
		StorageService:   *storageService,
		MediaInfoService: *mediaInfoService,
	}
}

func (c CourseService) Fetch() ([]models.Course, error) {
	return c.DBRepository.Fetch()
}

func (c CourseService) FetchById(id string) (models.Course, error) {
	return c.DBRepository.FetchById(id)
}

func (c CourseService) Create(data requests.CreateCourseRequest) (interface{}, error) {

	material_length := len(data.Materials)
	material_files_length := len(data.Files)

	if material_length != material_files_length {
		return nil, errors.New("Total Material Data and Material Files is not match")
	}

	//1. Construct Course Model
	var course models.Course

	timeNow := time.Now()

	course.Name = data.Name
	course.UserID = data.UserID
	course.Description = data.Description
	course.IsReleased = *data.IsReleased
	course.Price = data.Price
	course.UpdatedAt = &timeNow
	course.CreatedAt = &timeNow
	course.DeletedAt = nil

	if *data.IsReleased {
		course.ReleasedAt = &timeNow
	}

	//2. Upload Video to AWS S3 Bucket
	var uploadedMaterialVideo []response.S3Response

	var err error

	random := strconv.Itoa(rand.Int())
	courseName := data.Name + "-" + random + "/"

	uploadedMaterialVideo, err = c.StorageService.Upload(data.Files, courseName)
	if err != nil {
		return nil, err
	}

	uploadedCourseThumbnail, err := c.StorageService.Upload(data.Image, courseName)
	if err != nil {
		return nil, err
	}

	//3. Replace Aws s3 url to Aws Cloudfront url
	var regex, _ = regexp.Compile(`https://acourse-course-videos.s3.ap-southeast-1.amazonaws.com`)
	cloudfrontUrl := "https://d1wvipirze38am.cloudfront.net"

	course.Image = regex.ReplaceAllString(uploadedCourseThumbnail[0].Filepath, cloudfrontUrl)

	//4. Construct Course Materials
	total_duration := 0

	for i := 0; i < len(data.Materials); i++ {

		foundMaterial := func(materialOrder int, collections []response.S3Response) *response.S3Response {
			for _, material := range collections {
				if material.Order == materialOrder {
					return &material
				}
			}
			return nil
		}(*data.Materials[i].Order, uploadedMaterialVideo)

		duration, err := c.MediaInfoService.GetVideoDuration(data.Files[i])
		if err != nil {
			return nil, err
		}

		_duration, err := strconv.Atoi(duration)
		if err != nil {
			return nil, err
		}

		course.Materials = append(course.Materials, models.Material{
			Name:        data.Materials[i].Name,
			Duration:    time.Duration(_duration / 1000),
			Description: data.Materials[i].Description,
			Order:       *data.Materials[i].Order,
			Url:         regex.ReplaceAllString(foundMaterial.Filepath, cloudfrontUrl),
			Key:         foundMaterial.Key,
			UpdatedAt:   &timeNow,
			CreatedAt:   &timeNow,
			DeletedAt:   nil,
		})

		total_duration += _duration / 1000
	}

	course.TotalDuration = time.Duration(total_duration)

	//5. Save Course Model to Database
	course_id, err := c.DBRepository.Create(course)
	if err != nil {
		return models.Course{}, err
	}

	course.ID = course_id

	return course, nil
}
