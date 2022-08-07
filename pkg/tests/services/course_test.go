package services

import (
	"acourse-course-service/pkg/contracts"
	"acourse-course-service/pkg/database"
	"acourse-course-service/pkg/http/requests"
	repositories "acourse-course-service/pkg/repositories/database"
	s3repo "acourse-course-service/pkg/repositories/storage"
	"acourse-course-service/pkg/services"
	"context"
	"github.com/joho/godotenv"
	mediainfo "github.com/zhulik/go_mediainfo"
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
)

func init() {

	//Load .Env
	err := godotenv.Load("../../../.env")
	if err != nil {
		panic(err)
	}

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

func TestUpdateCourse(t *testing.T) {

	var updateCourseRequest requests.UpdateCourseRequest
	updateCourseRequest.Name = "Kelas Java"

	_, err := courseService.Update(ctx, updateCourseRequest, "62ef26e378e839637a8c466d")
	if err != nil {
		t.Errorf("Update Failed %v", err.Error())
		return
	}
	t.Logf("Update Success")
}
