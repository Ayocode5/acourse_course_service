package models

import (
	"time"
)

type Course struct {
	ID            string        `json:"id"`
	UserID        int64         `json:"user_id" bson:"user_id"`
	Name          string        `json:"name" bson:"name"`
	Description   string        `json:"description" bson:"description"`
	Image         string        `json:"image" bson:"image"`
	Price         float32       `json:"price" bson:"price"`
	TotalDuration time.Duration `json:"total_duration" bson:"total_duration"`
	IsReleased    bool          `json:"is_released" bson:"is_released"`
	Materials     []Material    `json:"materials" bson:"materials"`
	ReleasedAt    *time.Time    `json:"released_at,omitempty" bson:"released_at"`
	UpdatedAt     *time.Time    `json:"updated_at,omitempty" bson:"updated_at"`
	CreatedAt     *time.Time    `json:"created_at,omitempty" bson:"created_at"`
	DeletedAt     *time.Time    `json:"deleted_at,omitempty" bson:"deleted_at"`
}

type Material struct {
	Name        string        `json:"name" bson:"name"`
	Duration    time.Duration `json:"duration" bson:"duration"`
	Description string        `json:"description" bson:"description"`
	Order       int           `json:"order" bson:"order"`
	Url         string        `json:"url" bson:"url"`
	Key         string        `json:"key" bson:"key"`
	UpdatedAt   *time.Time    `json:"updated_at,omitempty" bson:"updated_at"`
	CreatedAt   *time.Time    `json:"created_at,omitempty" bson:"created_at"`
	DeletedAt   *time.Time    `json:"deleted_at,omitempty" bson:"deleted_at"`
}
