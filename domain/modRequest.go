package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type ModRequest struct {
	Id primitive.ObjectID `bson:"_id" json:"id"`
	ThreadOwnerUsername string `bson:"threadOwnerUsername" json:"threadOwnerUsername"`
	ModCandidateUsername string `bson:"modCandidateUsername" json:"modCandidateUsername"`
	OwnerAccept bool `bson:"ownerAccept" json:"ownerAccept"`
	CandidateAccept bool `bson:"candidateAccept" json:"candidateAccept"`
}
