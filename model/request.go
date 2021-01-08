package model

// For these structs, as in the case of my implementation there is no difference, bar the ID
// between the requests and the user object, these aren't used. If the user object contained
// additional internal parameters these would be used, but they're not here

// AddRequest is the request body expected to add a new user
type AddRequest struct {
	Forename string `json:"forename" dynamodbav:"forename"`
	Surname  string `json:"surname" dynamodbav:"surname"`
	Nickname string `json:"nickname" dynamodbav:"nickname"`
	Password string `json:"password" dynamodbav:"nickname"`
	Email    string `json:"email" dynamodbav:"email"`
	Country  string `json:"country" dynamodbav:"country"`
}

// UpdateRequest is the request body expected for an update operation
// kept separate in case certain fields need to be designated as unupdateable
type UpdateRequest struct {
	Forename string `json:"forename" dynamodbav:"forename"`
	Surname  string `json:"surname" dynamodbav:"surname"`
	Nickname string `json:"nickname" dynamodbav:"nickname"`
	Password string `json:"password" dynamodbav:"nickname"`
	Email    string `json:"email" dynamodbav:"email"`
	Country  string `json:"country" dynamodbav:"country"`
}
