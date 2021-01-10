package model

// FilterResponse is the struct returned by a user search
type FilterResponse struct {
	Results []*User `json:"results"`
	Count   int     `json:"count"`
}
