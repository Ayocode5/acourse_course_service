package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
	"time"
)

type Course struct {
	sync.Mutex
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	UserID        int64              `json:"user_id,omitempty" bson:"user_id"`
	Name          string             `json:"name,omitempty" bson:"name"`
	CourseID      string             `json:"course_id,omitempty" bson:"course_id"`
	Description   string             `json:"description,omitempty" bson:"description"`
	ImageUrl      string             `json:"image_url,omitempty" bson:"image_url"`
	ImageKey      string             `json:"image_key,omitempty" bson:"image_key"`
	Price         float32            `json:"price,omitempty" bson:"price"`
	TotalDuration time.Duration      `json:"total_duration,omitempty" bson:"total_duration"`
	IsReleased    bool               `json:"is_released,omitempty" bson:"is_released"`
	Materials     []Material         `json:"materials,omitempty" bson:"materials"`
	ReleasedAt    *time.Time         `json:"released_at,omitempty" bson:"released_at"`
	UpdatedAt     *time.Time         `json:"updated_at,omitempty" bson:"updated_at"`
	CreatedAt     *time.Time         `json:"created_at,omitempty" bson:"created_at"`
	DeletedAt     *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at"`
}

type Material struct {
	MaterialID  primitive.ObjectID `json:"material_id" bson:"material_id"`
	Name        string             `json:"name" bson:"name"`
	Duration    time.Duration      `json:"duration" bson:"duration"`
	Description string             `json:"description" bson:"description"`
	Order       int                `json:"order" bson:"order"`
	Url         string             `json:"url" bson:"url"`
	Key         string             `json:"key" bson:"key"`
	UpdatedAt   *time.Time         `json:"updated_at,omitempty" bson:"updated_at"`
	CreatedAt   *time.Time         `json:"created_at,omitempty" bson:"created_at"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at"`
}

type Pagination struct {
	Page    int64
	PerPage int64
}

func (p Pagination) GetPagination() (limit int64, skip int64) {
	return p.PerPage, (p.Page - 1) * p.PerPage
}

func (c *Course) AddTotalDuration(duration time.Duration) {
	c.Lock()
	c.TotalDuration += duration
	c.Unlock()
}

func (c *Course) SubTotalDuration(duration time.Duration) {
	c.Lock()
	c.TotalDuration -= duration
	c.Unlock()
}
