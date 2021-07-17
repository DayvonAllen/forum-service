package repo

import (
	"context"
	"example.com/app/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ThreadRepo interface {
	FindAll(string, context.Context) (*domain.Forum, error)
	Create(thread *domain.CreateThreadDto) error
	DeleteByID(primitive.ObjectID, string) error
}
