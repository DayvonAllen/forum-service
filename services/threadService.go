package services

import (
	"context"
	"example.com/app/domain"
	"example.com/app/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ThreadService interface {
	GetAllThreads(string, context.Context) (*domain.Forum, error)
	Create(thread *domain.Thread) error
	DeleteByID(primitive.ObjectID, string) error
}

type DefaultThreadService struct {
	repo repo.ThreadRepo
}

func (s DefaultThreadService) GetAllThreads(page string, ctx context.Context) (*domain.Forum, error) {
	u, err := s.repo.FindAll(page, ctx)
	if err != nil {
		return nil, err
	}
	return  u, nil
}

func (s DefaultThreadService) Create(thread *domain.Thread) error {
	err := s.repo.Create(thread)
	if err != nil {
		return err
	}
	return  nil
}

func (s DefaultThreadService) DeleteByID(id primitive.ObjectID, username string) error {
	err := s.repo.DeleteByID(id, username)
	if err != nil {
		return err
	}
	return nil
}

func NewThreadService(repository repo.ThreadRepo) DefaultThreadService {
	return DefaultThreadService{repository}
}