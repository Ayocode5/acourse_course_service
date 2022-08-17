package requests

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mime/multipart"
)

type CreateCourseRequest struct {
	UserID      int64                   `form:"user_id" json:"user_id" binding:"required"`
	Name        string                  `form:"name" json:"name" binding:"required"`
	Description string                  `form:"description" json:"description" binding:"required"`
	Price       float32                 `form:"price" json:"price" binding:"required"`
	IsReleased  *bool                   `form:"is_released" json:"is_released" binding:"required"`
	Materials   []CreateMaterialRequest `form:"materials" json:"materials" binding:"required,dive"`
	Files       []*multipart.FileHeader `form:"files" json:"files" binding:"required"`
	Image       *multipart.FileHeader   `form:"image" json:"image" binding:"required"`
}

type CreateMaterialRequest struct {
	MaterialID  primitive.ObjectID `json:"material_id"`
	Name        string             `form:"name" json:"name" binding:"required"`
	Description string             `form:"description" json:"description" binding:"required"`
	Order       *int               `form:"order" json:"order" binding:"required"`
	NewOrder    *int               `form:"new_order" json:"new_order"`
}

func (r CreateCourseRequest) ValidateMaterialFiles() error {

	material_length := len(r.Materials)
	material_files_length := len(r.Files)

	if material_length != material_files_length {
		return errors.New("Total Material Data and Material Files is not match")
	}

	return nil
}

type UpdateCourseRequest struct {
	UserID      int64                   `form:"user_id" json:"user_id"`
	Name        string                  `form:"name" json:"name"`
	Description string                  `form:"description" json:"description"`
	Price       *float32                `form:"price" json:"price"`
	IsReleased  *bool                   `form:"is_released" json:"is_released"`
	Materials   []CreateMaterialRequest `form:"materials" json:"materials"`
	Files       []*multipart.FileHeader `form:"files" json:"files"`
	Image       []*multipart.FileHeader `form:"image" json:"image"`
}

func (r UpdateCourseRequest) ValidateMaterialFiles() error {

	material_length := len(r.Materials)
	material_files_length := len(r.Files)

	if material_length != material_files_length {
		return errors.New("Total Material Data and Material Files is not match")
	}

	return nil
}

type DeleteMaterialsRequest struct {
	MaterialID []string `form:"material_id" json:"material_id" binding:"required"`
}
