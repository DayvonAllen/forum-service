package repo

import (
	"example.com/app/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type PostRepo interface {
	Create(reply *domain.Post) error
	FindAllPostsByResourceId(id primitive.ObjectID, username string, page string) ([]domain.Post, error)
	UpdateById(id primitive.ObjectID, newContent string, edited bool, updatedTime time.Time, username string) error
	LikePostById(primitive.ObjectID, string) error
	DisLikePostById(primitive.ObjectID, string) error
	DeleteById(id primitive.ObjectID, username string) error
	DeleteManyById(id primitive.ObjectID, username string) error
}

