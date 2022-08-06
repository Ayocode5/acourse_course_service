package requests

import "mime/multipart"

type CreateCourseRequest struct {
	UserID      int64                   `form:"user_id" json:"user_id" binding:"required"`
	Name        string                  `form:"name" json:"name" binding:"required"`
	Description string                  `form:"description" json:"description" binding:"required"`
	Price       float32                 `form:"price" json:"price" binding:"required"`
	IsReleased  *bool                   `form:"is_released" json:"is_released" binding:"required"`
	Materials   []CreateMaterialRequest `form:"materials" json:"materials" binding:"required,dive"`
	Files       []*multipart.FileHeader `form:"files" json:"files" binding:"required"`
	Image       []*multipart.FileHeader `form:"image" json:"image" binding:"required"`
}

type CreateMaterialRequest struct {
	Name        string `form:"name" json:"name" binding:"required"`
	Description string `form:"description" json:"description" binding:"required"`
	Order       *int   `form:"order" json:"order" binding:"required"`
}
