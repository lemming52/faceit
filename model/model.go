package model

import "time"

const (
	// UserAdd is the operation designation for messaging of adding a new user
	UserAdd = "AddNewUser"
	// UserDelete is the operation designation for messaging of removing user
	UserDelete = "DeleteUser"
	// UserUpdate is the operation designation for messaging of updating a new user
	UserUpdate = "UpdateUser"
)

// User is the major structure for the service, containing all required info and a unique key
type User struct {
	Id       string `json:"userId" dynamodbav:"userId"`
	Forename string `json:"forename" dynamodbav:"forename"`
	Surname  string `json:"surname" dynamodbav:"surname"`
	Nickname string `json:"nickname" dynamodbav:"nickname"`
	Password string `json:"password" dynamodbav:"password"`
	Email    string `json:"email" dynamodbav:"email"`
	Country  string `json:"country" dynamodbav:"country"`
}

// Message is the format of the messages emitted by the service
type Message struct {
	Id      string    `json:"userId"`
	Action  string    `json:"userAction"`
	Created time.Time `json:"creationTime"`
}

// New message converts a user Id and operation to a message
func NewMessage(id, action string) *Message {
	currentTime := time.Now()
	return &Message{
		Id:      id,
		Action:  action,
		Created: currentTime,
	}
}

// FilterCondition is a struct to siplify the transfer of query values between the service and the storage client
type FilterCondition struct {
	Query string
	Value interface{}
}
