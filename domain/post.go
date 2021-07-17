package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Post struct {
	Id             primitive.ObjectID `bson:"_id" json:"id"`
	AuthorUsername string			   `bson:"authorUsername" json:"authorUsername"`
	Content		string					`bson:"content" json:"content"`
	Edited 		bool					`bson:"edited" json:"edited"`
	Likes          []string           `bson:"likes" json:"-"`
	Dislikes       []string           `bson:"dislikes" json:"-"`
	LikeCount      int                `bson:"likeCount" json:"-"`
	DislikeCount   int                `bson:"dislikeCount" json:"-"`
	CreatedAt      time.Time          `bson:"createdAt" json:"-"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"-"`
}
