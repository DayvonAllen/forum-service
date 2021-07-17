package repo

import (
	"example.com/app/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ReplyRepo interface {
	Create(reply *domain.CreateReply) error
	FindAllRepliesByResourceId(id primitive.ObjectID, username string) (*[]domain.Reply, error)
	UpdateById(id primitive.ObjectID, newContent string, edited bool, updatedTime time.Time, username string) error
	LikeReplyById(primitive.ObjectID, string) error
	DisLikeReplyById(primitive.ObjectID, string) error
	DeleteById(id primitive.ObjectID, username string) error
}

