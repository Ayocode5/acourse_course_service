package services

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/http/middleware"
	"acourse-course-service/pkg/http/requests"
	"acourse-course-service/pkg/http/response"
	"acourse-course-service/pkg/models"
	"context"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
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

func (c CourseService) AuthorizeResourceByUserId(model *models.Course, auth middleware.Authorization) (bool, error) {

	userId, err := strconv.ParseInt(auth.UserID, 10, 64)
	if err != nil {
		return false, err
	}

	if model.UserID != userId {
		return false, nil
	}
	return true, nil
}

func (c CourseService) Fetch(ctx context.Context, excludeFields []string) ([]models.Course, error) {
	return c.DBRepository.Fetch(ctx, excludeFields)
}

func (c CourseService) FetchById(ctx context.Context, id string, excludeFields []string) (models.Course, error) {
	return c.DBRepository.FetchById(ctx, id, excludeFields)
}

func (c CourseService) Create(ctx context.Context, request requests.CreateCourseRequest) (interface{}, error) {

	//0. Validate Total Material & Files, if it's not match then return error
	validationErr := request.ValidateMaterialFiles()
	if validationErr != nil {
		return nil, validationErr
	}

	//Check if user id is not duplcate
	res, _ := c.DBRepository.FetchByUserId(ctx, request.UserID, []string{})
	if res.UserID == request.UserID {
		return nil, errors.New("duplicated user id")
	}

	//1. Construct Course Model
	var course models.Course

	timeNow := time.Now()
	course.ID = c.DBRepository.GenerateModelID()
	course.Name = request.Name
	course.UserID = request.UserID
	course.Description = request.Description
	course.IsReleased = *request.IsReleased
	course.Price = request.Price
	course.UpdatedAt = &timeNow
	course.CreatedAt = &timeNow
	course.DeletedAt = nil

	if *request.IsReleased {
		course.ReleasedAt = &timeNow
	}

	course.CourseID = request.Name + "-" + strconv.FormatInt(request.UserID, 10)

	//2. UploadFiles Video to AWS S3 Bucket
	var uploadedMaterialVideo []response.S3Response
	var err error

	uploadedMaterialVideo, err = c.StorageService.UploadFiles(request.Files, course.CourseID+"/")
	if err != nil {
		return nil, err
	}

	uploadedCourseThumbnail, err := c.StorageService.UploadFile(request.Image, course.CourseID+"/")
	if err != nil {
		return nil, err
	}
	course.ImageKey = uploadedCourseThumbnail.Key
	//3. Replace Aws s3 url to Aws Cloudfront url
	course.ImageUrl = c.replaceVideoUrl(uploadedCourseThumbnail.Filepath)

	//4. Construct Course Materials
	total_duration := 0

	for i := 0; i < len(request.Materials); i++ {

		existingMaterial := func(materialOrder int, collections []response.S3Response) *response.S3Response {
			for _, material := range collections {
				if material.Order == materialOrder {
					return &material
				}
			}
			return nil
		}(*request.Materials[i].Order, uploadedMaterialVideo)

		//Get Video Duration
		duration, _ := c.getVideoDuration(request.Files[i])

		course.Materials = append(course.Materials, models.Material{
			MaterialID:  c.DBRepository.GenerateModelID(),
			Name:        request.Materials[i].Name,
			Duration:    time.Duration(duration),
			Description: request.Materials[i].Description,
			Order:       *request.Materials[i].Order,
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
	courseId, err := c.DBRepository.Create(ctx, &course)
	if err != nil {
		return nil, err
	}

	course.ID = courseId

	return course, nil
}

func (c CourseService) Update(ctx context.Context, request requests.UpdateCourseRequest, courseId string) (*response.HttpResponse, error) {

	authorization := ctx.Value("authorization").(*middleware.Authorization)

	//Fetch Course By id
	var course models.Course

	course, err := c.DBRepository.FetchById(ctx, courseId, []string{})
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, err
	}

	//Authorize User to check the rights to do manipulation
	validated, err := c.AuthorizeResourceByUserId(&course, *authorization)
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, nil
	}

	if !validated {
		return &response.HttpResponse{
			StatusCode: http.StatusUnauthorized,
			Message:    "You don't have any permission to edit this resources",
		}, nil
	}

	//Update current request
	timeNow := time.Now()

	if request.Name != "" {
		course.Name = request.Name
	}
	if request.Description != "" {
		course.Description = request.Description
	}
	if request.IsReleased != nil {
		course.IsReleased = *request.IsReleased
	}
	if request.Price != nil {
		course.Price = *request.Price
	}
	course.UpdatedAt = &timeNow

	//Update Current Materials
	var wg sync.WaitGroup

	for i := 0; i < len(request.Materials); i++ {

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

					log.Println("updating material request is done")
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
		}(&wg, &course, request, i)
	}

	wg.Wait()

	_, err = c.DBRepository.Update(ctx, course, courseId)
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, nil
	}

	return &response.HttpResponse{
		StatusCode: http.StatusOK,
		Message:    "Updated successfully",
	}, nil
}

func (c CourseService) DeleteMaterials(ctx context.Context, course_id string, materialIds []string) (*response.HttpResponse, error) {

	authorization := ctx.Value("authorization").(*middleware.Authorization)

	course, err := c.DBRepository.FetchById(ctx, course_id, []string{})
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, err
	}

	validated, err := c.AuthorizeResourceByUserId(&course, *authorization)
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, nil
	}

	if !validated {
		return &response.HttpResponse{
			StatusCode: http.StatusUnauthorized,
			Message:    "You don't have any permission to edit this resources",
		}, nil
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
			_ = c.StorageService.Delete(foundMaterial.Key)
			//if err != nil {
			//	return false, err
			//}

			//Decrease Duration
			course.SubTotalDuration(foundMaterial.Duration)
		}
	}

	_, err = c.DBRepository.Update(ctx, course, course_id)
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, err
	}

	_, err = c.DBRepository.DeleteMaterials(ctx, course_id, materialIds)
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, err
	}

	return &response.HttpResponse{
		StatusCode: http.StatusOK,
		Message:    "Meterial deleted successfully",
	}, nil
}

func (c CourseService) DeleteCourse(ctx context.Context, course_id string) (*response.HttpResponse, error) {

	authorization := ctx.Value("authorization").(*middleware.Authorization)

	//1. Fetch Course
	course, err := c.DBRepository.FetchById(ctx, course_id, []string{
		"name", "image", "price", "description", "created_at", "updated_at", "is_released", "course_id", "total_duration", "user_id"})
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, err
	}

	validated, err := c.AuthorizeResourceByUserId(&course, *authorization)
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, nil
	}

	if !validated {
		return &response.HttpResponse{
			StatusCode: http.StatusUnauthorized,
			Message:    "You don't have any permission to edit this resources",
		}, nil
	}

	//2. Delete All Material Video From Storage Service
	for _, material := range course.Materials {
		_ = c.StorageService.Delete(material.Key)
		//if err != nil {
		//	return false, err
		//}
	}

	//3. Delete Course's Thumbnail
	_ = c.StorageService.Delete(course.ImageKey)
	//if err != nil {
	//	return false, err
	//}

	//4. Delete Course Data from Database
	_, err = c.DBRepository.DeleteCourse(ctx, course_id)
	if err != nil {
		return &response.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}, err
	}

	return &response.HttpResponse{
		StatusCode: http.StatusOK,
		Message:    "Course deleted successfully",
	}, nil

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
