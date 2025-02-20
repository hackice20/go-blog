package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Blog struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Title     string               `bson:"title" json:"title"`
	Content   string               `bson:"content" json:"content"`
	Author    primitive.ObjectID   `bson:"author" json:"author"`
	Likes     []primitive.ObjectID `bson:"likes" json:"likes"`
	Comments  []Comment            `bson:"comments" json:"comments"`
	CreatedAt time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time            `bson:"updated_at" json:"updated_at"`
}

type Comment struct {
	Text      string             `bson:"text" json:"text"`
	Author    primitive.ObjectID `bson:"author" json:"author"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
