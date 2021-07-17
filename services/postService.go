package services

import (
	"example.com/app/domain"
	"example.com/app/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type PostService interface {
	Create(post *domain.Post) error
	FindAllPostsByResourceId(id primitive.ObjectID,  username string) (*[]domain.Post, error)
	UpdateById(id primitive.ObjectID, newContent string, edited bool, updatedTime time.Time, username string) error
	LikeCommentById(primitive.ObjectID, string) error
	DisLikeCommentById(primitive.ObjectID, string) error
	DeleteById(id primitive.ObjectID, username string) error
}

type DefaultPostService struct {
	repo repo.PostRepo
}

func (c DefaultPostService) Create(post *domain.Post) error {
	err := c.repo.Create(post)
	if err != nil {
		return err
	}
	return nil
}

func (c DefaultPostService) FindAllPostsByResourceId(id primitive.ObjectID,  username string) (*[]domain.Post, error) {
	comment, err := c.repo.FindAllPostsByResourceId(id, username)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (c DefaultPostService) UpdateById(id primitive.ObjectID, newContent string, edited bool, updatedTime time.Time, username string) error {
	err := c.repo.UpdateById(id, newContent, edited, updatedTime, username)
	if err != nil {
		return err
	}
	return nil
}

func (c DefaultPostService) LikeCommentById(id primitive.ObjectID, username string) error {
	err := c.repo.LikePostById(id, username)
	if err != nil {
		return err
	}
	return nil
}

func (c DefaultPostService) DisLikeCommentById(id primitive.ObjectID, username string) error {
	err := c.repo.DisLikePostById(id, username)
	if err != nil {
		return err
	}
	return nil
}

func (c DefaultPostService) DeleteById(id primitive.ObjectID, username string) error {
	err := c.repo.DeleteById(id, username)
	if err != nil {
		return err
	}
	return nil
}

func NewPostService(repository repo.PostRepo) DefaultPostService {
	return DefaultPostService{repository}
}