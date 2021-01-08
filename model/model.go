package model

import "time"

const (
	UserAdd    = "AddNewUser"
	UserDelete = "DeleteUser"
	UserUpdate = "UpdateUser"
)

type User struct {
	Id       string `json:"userId" dynamodbav:"userId"`
	Forename string `json:"forename" dynamodbav:"forename"`
	Surname  string `json:"surname" dynamodbav:"surname"`
	Nickname string `json:"nickname" dynamodbav:"nickname"`
	Password string `json:"password" dynamodbav:"password"`
	Email    string `json:"email" dynamodbav:"email"`
	Country  string `json:"country" dynamodbav:"country"`
}

type Message struct {
	Id      string    `json:"userId"`
	Action  string    `json:"userAction"`
	Created time.Time `json:"creationTime"`
}

func NewMessage(id, action string) *Message {
	currentTime := time.Now()
	return &Message{
		Id:      id,
		Action:  action,
		Created: currentTime,
	}
}

type FilterCondition struct {
	Query string
	Value interface{}
}
