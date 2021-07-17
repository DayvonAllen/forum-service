package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Thread struct {
	Id primitive.ObjectID 		`bson:"_id" json:"id"`
	OwnerUsername string  		`bson:"ownerUsername" json:"-"`
	Name  string				`bson:"name" json:"name"`
	Description string			`bson:"description" json:"description"`
	Posts []Post				`bson:"-" json:"posts"`
	NumberOfPosts int			`bson:"numberOfPosts" json:"numberOfPosts"`
	Score int					`bson:"score" json:"-"`
	Mods []string				`bson:"mods" json:"-"`
	Banned []string				`bson:"banned" json:"-"`
	DisableModRequest bool		`bson:"disableModRequest" json:"disableModRequest"`
	CreatedAt   time.Time		`bson:"createdAt" json:"-"`
	UpdatedAt   time.Time		`bson:"updatedAt" json:"-"`
}

type CreateThreadDto struct {
	OwnerUsername string  		`bson:"ownerUsername" json:"-"`
	Name  string				`bson:"name" json:"name"`
	Description string			`bson:"description" json:"description"`
}

type ThreadPreview struct {
	Id primitive.ObjectID 		`bson:"_id" json:"-"`
	Name  string				`bson:"name" json:"name"`
	Description string			`bson:"description" json:"description"`
	NumberOfPosts int			`bson:"numberOfPosts" json:"numberOfPosts"`
	CreatedAt   time.Time		`bson:"createdAt" json:"-"`
	UpdatedAt   time.Time		`bson:"updatedAt" json:"-"`
}

type FindThreadDto struct {
	Name  string				`bson:"name" json:"name"`
}
