package models

type Course struct {
	Name      string     `bson:"name"`
	Materials []Material `bson:"materials"`
}

type Material struct {
	Name string `bson:"name"`
}
