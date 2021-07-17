package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Post struct {
	Id             primitive.ObjectID `bson:"_id" json:"-"`
	ResourceId     primitive.ObjectID `bson:"resourceId" json:"-"`
	Content        string             `bson:"content" json:"content"`
	AuthorUsername string             `bson:"authorUsername" json:"authorUsername"`
	Edited         bool               `bson:"edited" json:"edited"`
	Likes          []string           `bson:"likes" json:"-"`
	Dislikes       []string           `bson:"dislikes" json:"-"`
	LikeCount      int                `bson:"likeCount" json:"likeCount"`
	DislikeCount   int                `bson:"dislikeCount" json:"dislikeCount"`
	CurrentUserLiked    bool          `bson:"-" json:"currentUserLiked"`
	CurrentUserDisLiked bool          `bson:"-" json:"currentUserDisLiked"`
	Replies             *[]Reply      `bson:"-" json:"replies"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type CreatePost struct {
	ResourceId     primitive.ObjectID `bson:"resourceId" json:"-"`
	Content        string             `bson:"content" json:"content"`
	AuthorUsername string             `bson:"authorUsername" json:"-"`
}