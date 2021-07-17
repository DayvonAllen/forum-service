package domain

type Forum struct {
	Threads *[]Thread       `bson:"threads" json:"threads"`
}
