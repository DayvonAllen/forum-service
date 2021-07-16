package domain

// Message messageType 201 user created
// messageType 200 user updated
type Message struct {
	User         User   `form:"User" json:"User"`
	MessageType  int    `form:"messageType" json:"messageType"`
	ResourceType string `form:"resourceType" json:"resourceType"`
}