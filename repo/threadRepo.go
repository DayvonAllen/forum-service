package repo

import (
	"context"
	"example.com/app/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ThreadRepo interface {
	FindAll(string, context.Context) (*[]domain.ThreadPreview, error)
	FindByName(string, string, string) (*domain.Thread, error)
	Create(thread *domain.Thread) error
	DeleteByID(primitive.ObjectID, string) error
}
