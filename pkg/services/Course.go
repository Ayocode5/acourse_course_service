package services

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/requests"
	"acourse-course-service/pkg/http/response"
	"acourse-course-service/pkg/models"
	"context"
	"errors"
	"log"
	"mime/multipart"
	"regexp"
	"strconv"
	"sync"
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

func (c CourseService) Fetch(ctx context.Context, excludeFields []string) ([]models.Course, error) {
	return c.DBRepository.Fetch(ctx, excludeFields)
}

func (c CourseService) FetchById(ctx context.Context, id string, excludeFields []string) (models.Course, error) {
	return c.DBRepository.FetchById(ctx, id, excludeFields)
}

func (c CourseService) Create(ctx context.Context, data requests.CreateCourseRequest) (interface{}, error) {

	//0. Validate Total Material & Files, if it's not match then return error
	validationErr := data.ValidateMaterialFiles()
	if validationErr != nil {
		return nil, validationErr
	}

	//Check if user id is not duplcate
	res, _ := c.DBRepository.FetchByUserId(ctx, data.UserID)
	if res.UserID == data.UserID {
		return nil, errors.New("duplicated user id")
	}

	//1. Construct Course Model
	var course models.Course

	timeNow := time.Now()
	course.ID = c.DBRepository.GenerateModelID()
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

	course.CourseID = data.Name + "-" + strconv.FormatInt(data.UserID, 10)

	//2. UploadFiles Video to AWS S3 Bucket
	var uploadedMaterialVideo []response.S3Response
	var err error

	uploadedMaterialVideo, err = c.StorageService.UploadFiles(data.Files, course.CourseID+"/")
	if err != nil {
		return nil, err
	}

	uploadedCourseThumbnail, err := c.StorageService.UploadFile(data.Image, course.CourseID+"/")
	if err != nil {
		return nil, err
	}
	course.ImageKey = uploadedCourseThumbnail.Key
	//3. Replace Aws s3 url to Aws Cloudfront url
	course.ImageUrl = c.replaceVideoUrl(uploadedCourseThumbnail.Filepath)

	//4. Construct Course Materials
	total_duration := 0

	for i := 0; i < len(data.Materials); i++ {

		existingMaterial := func(materialOrder int, collections []response.S3Response) *response.S3Response {
			for _, material := range collections {
				if material.Order == materialOrder {
					return &material
				}
			}
			return nil
		}(*data.Materials[i].Order, uploadedMaterialVideo)

		//Get Video Duration
		duration, _ := c.getVideoDuration(data.Files[i])

		course.Materials = append(course.Materials, models.Material{
			MaterialID:  c.DBRepository.GenerateModelID(),
			Name:        data.Materials[i].Name,
			Duration:    time.Duration(duration),
			Description: data.Materials[i].Description,
			Order:       *data.Materials[i].Order,
			Url:         c.replaceVideoUrl(existingMaterial.Filepath),
			Key:         existingMaterial.Key,
			UpdatedAt:   &timeNow,
			CreatedAt:   &timeNow,
			DeletedAt:   nil,
		})

		total_duration += duration
	}

	course.TotalDuration = time.Duration(total_duration)

	//5. Save Course Model to Database
	courseId, err := c.DBRepository.Create(ctx, course)
	if err != nil {
		return nil, err
	}

	course.ID = courseId

	return course, nil
}

func (c CourseService) Update(ctx context.Context, data requests.UpdateCourseRequest, courseId string) (bool, error) {

	var wg sync.WaitGroup

	//Fetch Course By id
	var course models.Course

	course, err := c.DBRepository.FetchById(ctx, courseId, []string{})
	if err != nil {
		return false, err
	}

	//Update current data
	timeNow := time.Now()

	if data.Name != "" {
		course.Name = data.Name
	}
	if data.Description != "" {
		course.Description = data.Description
	}
	if data.IsReleased != nil {
		course.IsReleased = *data.IsReleased
	}
	if data.Price != nil {
		course.Price = *data.Price
	}
	course.UpdatedAt = &timeNow

	//Update Current Materials
	for i := 0; i < len(data.Materials); i++ {

		wg.Add(1)

		go func(wg *sync.WaitGroup, course *models.Course, data requests.UpdateCourseRequest, i int) {
			defer wg.Done()

			//Check Material if exists
			existingMaterial := func(materialID interface{}, course *models.Course) *models.Material {
				for i := 0; i < len(course.Materials); i++ {
					if materialID == course.Materials[i].MaterialID {
						return &course.Materials[i]
					}
				}
				return nil
			}(data.Materials[i].MaterialID, course)

			//If Material Exists Update the Material, Otherwise Add New one
			if existingMaterial != nil {

				log.Println("updating existing material")
				//Reset Material Ordering
				materialOrder := data.Materials[i].Order
				if data.Materials[i].NewOrder != nil {
					materialOrder = data.Materials[i].NewOrder
				}

				//Assign New Data to existing material
				existingMaterial.Name = data.Materials[i].Name
				existingMaterial.Description = data.Materials[i].Description
				existingMaterial.Order = *materialOrder
				existingMaterial.UpdatedAt = &timeNow

				//Update Material Video if exists on request file
				if len(data.Files) >= (i + 1) {

					log.Println("replacing old video")

					//UploadFiles new Video
					uploadedNewVideo, err := c.StorageService.UploadFile(data.Files[i], course.CourseID+"/")
					if err != nil {
						log.Fatal(err.Error())
						//return false, err
					}

					//Delete Old Video
					err = c.StorageService.Delete(existingMaterial.Key)
					if err != nil {
						log.Fatal(err.Error())
						//return false, err
					}

					//Decrease Course Total Duration by Material Duration
					course.SubTotalDuration(existingMaterial.Duration)
					//Renew Video Url
					existingMaterial.Url = c.replaceVideoUrl(uploadedNewVideo.Filepath)
					existingMaterial.Key = uploadedNewVideo.Key
					//Renew Duration Information
					duration, _ := c.getVideoDuration(data.Files[i])
					existingMaterial.Duration = time.Duration(duration)
					//Re-Adding Total Duration
					course.AddTotalDuration(time.Duration(duration))

					log.Println("updating material data is done")
				}

			} else {

				log.Println("adding new material")
				var err error
				var newVideoKey string
				var uploadedNewVideo response.S3Response
				var newVideoDuration int
				var cloudfrontVideoUrl string

				if len(data.Files) >= (i + 1) {

					log.Println("uploading new video")

					uploadedNewVideo, err = c.StorageService.UploadFile(data.Files[i], course.CourseID+"/")
					if err != nil {
						//return false, err
					}

					newVideoKey = uploadedNewVideo.Key

					newVideoDuration, _ = c.getVideoDuration(data.Files[i])
					cloudfrontVideoUrl = c.replaceVideoUrl(uploadedNewVideo.Filepath)
				}

				//Otherwise, Add New Material
				course.Materials = append(course.Materials, models.Material{
					MaterialID:  c.DBRepository.GenerateModelID(),
					Name:        data.Materials[i].Name,
					Order:       *data.Materials[i].Order,
					Description: data.Materials[i].Description,
					Key:         newVideoKey,
					Url:         cloudfrontVideoUrl,
					Duration:    time.Duration(newVideoDuration),
					UpdatedAt:   &timeNow,
					CreatedAt:   &timeNow,
				})

				course.AddTotalDuration(time.Duration(newVideoDuration))

				log.Println("adding new material is done")
			}

		}(&wg, &course, data, i)

	}

	wg.Wait()

	_, err = c.DBRepository.Update(ctx, course, courseId)
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	return true, nil
}

func (c CourseService) DeleteMaterials(ctx context.Context, course_id string, materialIds []string) (bool, error) {

	course, err := c.DBRepository.FetchById(ctx, course_id, []string{})
	if err != nil {
		return false, err
	}

	for _, m_id := range materialIds {

		foundMaterial := func(course models.Course, targetID string) *models.Material {
			for _, material := range course.Materials {
				if material.MaterialID.Hex() == targetID {
					return &material
				}
			}
			return nil
		}(course, m_id)

		if foundMaterial != nil {
			err := c.StorageService.Delete(foundMaterial.Key)
			if err != nil {
				return false, err
			}

			//Decrease Duration
			course.SubTotalDuration(foundMaterial.Duration)
		}
	}

	_, err = c.DBRepository.Update(ctx, course, course_id)
	if err != nil {
		return false, err
	}

	_, err = c.DBRepository.DeleteMaterials(ctx, course_id, materialIds)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c CourseService) DeleteCourse(ctx context.Context, course_id string) (bool, error) {

	course, err := c.DBRepository.FetchById(ctx, course_id, []string{
		"name", "image", "price", "description", "created_at", "updated_at", "is_released", "course_id", "total_duration", "user_id"})
	if err != nil {
		return false, err
	}

	for _, material := range course.Materials {
		err := c.StorageService.Delete(material.Key)
		if err != nil {
			return false, err
		}
	}

	err = c.StorageService.Delete(course.ImageKey)
	if err != nil {
		return false, err
	}

	//err = c.StorageService.Delete(course.CourseID + "/")
	//if err != nil {
	//	return false, err
	//}

	res, err := c.DBRepository.DeleteCourse(ctx, course_id)
	if err != nil {
		return false, err
	}

	return res, nil

}

func (s CourseService) replaceVideoUrl(url string) string {
	var regex, _ = regexp.Compile(`https://acourse-course-videos.s3.ap-southeast-1.amazonaws.com`)
	return regex.ReplaceAllString(url, "https://d1wvipirze38am.cloudfront.net")
}

func (s CourseService) getVideoDuration(file *multipart.FileHeader) (int, error) {

	duration, err := s.MediaInfoService.GetVideoDuration(file)
	if err != nil {
		return 0, err
	}

	_duration, err := strconv.Atoi(duration)
	if err != nil {
		return 0, err
	}

	return _duration / 1000, nil
}
