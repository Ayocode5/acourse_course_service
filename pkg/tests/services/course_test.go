package services

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/database"
	"acourse-course-service/pkg/http/controllers"
	"acourse-course-service/pkg/http/requests"
	repositories "acourse-course-service/pkg/repositories/database"
	s3repo "acourse-course-service/pkg/repositories/storage"
	"acourse-course-service/pkg/services"
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	mediainfo "github.com/zhulik/go_mediainfo"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	db                  database.Database
	dbRepository        contracts.CourseDatabaseRepository
	s3StorageRepository contracts.StorageRepository
	storageService      contracts.StorageService
	mediaInfoService    contracts.MediaInfoService
	courseService       contracts.CourseService
	ctx                 context.Context
	engine              *gin.Engine
)

func init() {

	ctx = context.Background()

	//Load .Env
	err := godotenv.Load("../../../.env")
	if err != nil {
		panic(err)
	}

	//Create Gin Instance
	engine = gin.Default()

	db = database.Database{
		DbName:       os.Getenv("DB_NAME"),
		DbCollection: os.Getenv("DB_COLLECTION"),
		DbHost:       os.Getenv("DB_HOST"),
		DbPort:       os.Getenv("DB_PORT"),
	}

	//Open Connection
	mongodb := db.Prepare()

	//Setup MongoDB Repository
	dbRepository = repositories.ConstructDBRepository(mongodb.GetConnection(), mongodb.GetCollection())

	//Setup S3 Storage Repository
	s3StorageRepository = s3repo.ConstructS3Repository(
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		os.Getenv("AWS_BUCKET_NAME"),
		os.Getenv("AWS_BUCKET_REGION"),
		[]string{"video/mp4", "video/x-matroska", "image/jpeg"},
		int64(100*1024*1024),
		3,
	)

	//Setup Storage CourseService
	storageService = services.ConstructStorageService(&s3StorageRepository)

	//Setup MediaInfo Service
	mediaInfoService = services.ConstructMediaInfoService(mediainfo.NewMediaInfo())

	//Setup Course Services
	courseService = services.ConstructCourseService(&dbRepository, &storageService, &mediaInfoService)

	//Setup Course Devlivery/Http Controller
	controllers.SetupCourseHandler(ctx, engine, courseService)

}

func TestMaterialFilesNotMatch(t *testing.T) {

	var material1 requests.CreateMaterialRequest
	order1 := 0
	material1.Name = "Materi 1 Title"
	material1.Order = &order1
	material1.Description = "Materi 1 Description"

	var material2 requests.CreateMaterialRequest
	order2 := 1
	material2.Name = "Materi 2 Title"
	material2.Order = &order2
	material2.Description = "Materi 2 Description"

	var released *bool

	var createCourseRequest requests.CreateCourseRequest
	createCourseRequest.Name = "Kelas Go"
	createCourseRequest.Description = "Kelas Go Desc"
	createCourseRequest.UserID = 100
	createCourseRequest.Price = 99999
	createCourseRequest.IsReleased = released
	createCourseRequest.Materials = append(createCourseRequest.Materials, material1)
	createCourseRequest.Materials = append(createCourseRequest.Materials, material2)

	_, err := courseService.Create(ctx, createCourseRequest)
	if err != nil {
		t.Logf("Jumlah File dan Material data tidak sama: %v", err.Error())
	}
}

func TestCreateCourse(t *testing.T) {

	body := new(bytes.Buffer)

	mw := multipart.NewWriter(body)

	formField := map[string]string{
		"user_id":     "6872",
		"name":        "Kelas Python",
		"description": "Desc Kelas Python",
		"price":       "99999",
		"is_released": "true",
	}

	formFields := map[string][]string{
		"materials": {
			`{"name": "1. Intro", "description": "desc intro", "order": 0}`,
			`{"name": "2. Sejarah", "description": "desc sejarah", "order": 1}`,
		},
	}

	filePath := "/run/media/global/Data Y/Documents/Projects/Microservices/Go/acourse-course-service/pkg/tests/services/files/"

	formFiles := map[string][]string{
		"image": {filePath + "thumbnail.jpg"},
		"files": {filePath + "test1.mp4", filePath + "test2.mp4"},
	}

	for k, v := range formField {
		mw.WriteField(k, v)
	}

	for k, v := range formFields {
		for _, j := range v {
			mw.WriteField(k, j)
		}
	}

	for k, v := range formFiles {
		for _, j := range v {

			file, err := os.Open(j)
			if err != nil {
				t.Fatal(err)
			}

			w, err := mw.CreateFormFile(k, j)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := io.Copy(w, file); err != nil {
				t.Fatal(err)
			}
		}
	}

	// close the writer before making the request
	mw.Close()

	req, _ := http.NewRequest("POST", "/course/create", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())

	//Listen Result
	res := httptest.NewRecorder()

	//Procced Request
	engine.ServeHTTP(res, req)

	//Read Result
	response, _ := ioutil.ReadAll(res.Body)

	//Test Results
	//mockResponse := `{"message":"Course Successfuly Created"}`
	//assert.Equal(t, mockResponse, string(response))
	t.Log(string(response))
	assert.Equal(t, http.StatusCreated, res.Code)
}

func TestCantCreateDuplicateCourseUserId(t *testing.T) {

	body := new(bytes.Buffer)

	mw := multipart.NewWriter(body)

	formField := map[string]string{
		"user_id":     "99",
		"name":        "Kelas Python",
		"description": "Desc Kelas Python",
		"price":       "99999",
		"is_released": "true",
	}

	formFields := map[string][]string{
		"materials": {
			`{"name": "1. Intro", "description": "desc intro", "order": 0}`,
			`{"name": "2. Sejarah", "description": "desc sejarah", "order": 1}`,
		},
	}

	filePath := "/run/media/global/Data Y/Documents/Projects/Microservices/Go/acourse-course-service/pkg/tests/services/files/"

	formFiles := map[string][]string{
		"image": {filePath + "thumbnail.jpg"},
		"files": {filePath + "test1.mp4", filePath + "test2.mp4"},
	}

	for k, v := range formField {
		mw.WriteField(k, v)
	}

	for k, v := range formFields {
		for _, j := range v {
			mw.WriteField(k, j)
		}
	}

	for k, v := range formFiles {
		for _, j := range v {

			file, err := os.Open(j)
			if err != nil {
				t.Fatal(err)
			}

			w, err := mw.CreateFormFile(k, j)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := io.Copy(w, file); err != nil {
				t.Fatal(err)
			}
		}
	}

	// close the writer before making the request
	mw.Close()

	req, _ := http.NewRequest("POST", "/course/create", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())

	//Listen Result
	res := httptest.NewRecorder()

	//Procced Request
	engine.ServeHTTP(res, req)

	response, _ := ioutil.ReadAll(res.Body)
	t.Log(string(response))
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestUpdateCourseMaterial(t *testing.T) {

	filePath := "/run/media/global/Data Y/Documents/Projects/Microservices/Go/acourse-course-service/pkg/tests/services/files/"

	body := new(bytes.Buffer)

	mw := multipart.NewWriter(body)

	formField := map[string]string{
		"name":        "Kelas Java",
		"description": "Desc Kelas Java",
		"price":       "30000",
		"is_released": "true",
		//"image": filePath + "thumbnail.jpg",
	}

	courseID := "62f510437171ea7969fa9183"

	formFields := map[string][]string{
		"materials": {
			`{"material_id": "","name": "Thread Java Update", "description": "desc Thread", "order": 1}`,
			`{"material_id": "", "name": "Paralel Java Update", "description": "desc Concurent", "order": 2}`,
			`{"material_id": "", "name": "Concurent Java Update", "description": "desc Concurent", "order": 3}`,
		},
	}

	formFiles := map[string][]string{
		//"image": {filePath + "thumbnail.jpg"},
		"files": {
			filePath + "test1.mp4",
			filePath + "test2.mp4",
			//filePath + "test3.mp4",
			filePath + "test4.mp4",
		},
	}

	for k, v := range formField {
		mw.WriteField(k, v)
	}

	for k, v := range formFields {
		for _, j := range v {
			mw.WriteField(k, j)
		}
	}

	for k, v := range formFiles {
		for _, j := range v {

			file, err := os.Open(j)
			if err != nil {
				t.Fatal(err)
			}

			w, err := mw.CreateFormFile(k, j)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := io.Copy(w, file); err != nil {
				t.Fatal(err)
			}
		}
	}

	// close the writer before making the request
	mw.Close()

	req, _ := http.NewRequest("PUT", "/course/update/"+courseID, body)
	req.Header.Add("Content-Type", mw.FormDataContentType())

	//Listen Result
	res := httptest.NewRecorder()

	//Procced Request
	engine.ServeHTTP(res, req)

	//Test Results
	response, _ := ioutil.ReadAll(res.Body)
	t.Log(string(response))
	assert.Equal(t, http.StatusOK, res.Code)

}

func TestDeleteMaterials(t *testing.T) {
	res, err := courseService.DeleteMaterials(ctx, "62f52b92d51d3a9f6331d4c1", []string{"62f52b98d51d3a9f6331d4c2"})
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, true, res)
}

func TestDeleteCourse(t *testing.T) {
	res, err := courseService.DeleteCourse(ctx, "62f531ff5c0be0bd4ea7fdb2")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, res)
}
